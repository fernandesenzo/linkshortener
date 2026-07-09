package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/linkshortener/internal/link/codegen"
	"github.com/fernandesenzo/linkshortener/internal/link/handler"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
	"github.com/fernandesenzo/linkshortener/internal/link/service"
	"github.com/fernandesenzo/linkshortener/internal/middleware"
	"github.com/redis/go-redis/v9"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := repository.NewRedisRepository(client)
	svc := service.New(codegen.New(), repo)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.Handle("POST /api/links", middleware.BodyLimit(4096)(http.HandlerFunc(h.Create)))
	mux.HandleFunc("GET /{code}", h.Get)

	var handlerStack http.Handler = mux
	handlerStack = middleware.AccessLog(handlerStack)
	handlerStack = middleware.ApplyHeaders("*")(handlerStack)
	handlerStack = middleware.InjectReqID(handlerStack)
	handlerStack = middleware.Recover(handlerStack)

	return httptest.NewServer(handlerStack)
}

func TestCreateLink(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":"https://example.com"}`
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] == "" || len(result["code"]) != 6 {
		t.Errorf("expected code with length 6, got %q", result["code"])
	}
}

func TestRedirectValidCode(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":"https://example.com"}`
	resp, _ := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp2, err := client.Get(ts.URL + "/" + result["code"])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected 307, got %d", resp2.StatusCode)
	}
	if resp2.Header.Get("Location") != "https://example.com" {
		t.Errorf("expected Location https://example.com, got %q", resp2.Header.Get("Location"))
	}
}

func TestRedirectInvalidCode(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/ZZZZZZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateLinkInvalidURL(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":"not-a-valid-url"}`
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateLinkURLTooLong(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	longURL := "http://a" + strings.Repeat("b", 200) + ".com"
	body := `{"url":"` + longURL + `"}`
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestCreateLinkWrongContentType(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Post(ts.URL+"/api/links", "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", resp.StatusCode)
	}
}

func TestCreateLinkMalformedJSON(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader("{broken"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateLinkUnknownFields(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":"https://example.com","extra":"field"}`
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestIPRateLimit(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":"https://example.com"}`
	client := ts.Client()

	for i := 0; i < 10; i++ {
		resp, err := client.Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("request %d: expected 201, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp, err := client.Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 on 11th request, got %d", resp.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/ZZZZZZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", resp.Header.Get("X-Content-Type-Options"))
	}
}

func TestInvalidCodeLength(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/ab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid code length, got %d", resp.StatusCode)
	}
}

func TestUnknownRoute(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/api/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateLinkRequestBody(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestBodyTooLarge(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	bigBody := strings.Repeat("x", 5000)
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(bigBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
}

func TestCreateLinkEmptyURL(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	body := `{"url":""}`
	resp, err := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetLinkResponseBody(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	createBody := `{"url":"https://example.com"}`
	resp, _ := ts.Client().Post(ts.URL+"/api/links", "application/json", strings.NewReader(createBody))
	var createResult map[string]string
	json.NewDecoder(resp.Body).Decode(&createResult)
	resp.Body.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp2, err := client.Get(ts.URL + "/" + createResult["code"])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("Location") != "https://example.com" {
		t.Errorf("expected Location header to be https://example.com, got %q", resp2.Header.Get("Location"))
	}
}
