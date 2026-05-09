package codegen_test

import (
	"strings"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link/codegen"
)

func TestRandomGenerator_Generate(t *testing.T) {
	g := codegen.New()

	t.Run("should generate code with correct length", func(t *testing.T) {
		lengths := []int{1, 5, 10, 20}
		for _, length := range lengths {
			code, err := g.Generate(length)
			if err != nil {
				t.Errorf("unexpected error for length %d: %v", length, err)
			}
			if len(code) != length {
				t.Errorf("expected length %d, got %d", length, len(code))
			}
		}
	})

	t.Run("should only contain valid characters", func(t *testing.T) {
		charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		code, _ := g.Generate(100)
		for _, char := range code {
			if !strings.ContainsRune(charset, char) {
				t.Errorf("generated code contains invalid character: %c", char)
			}
		}
	})

	t.Run("should return error for invalid length", func(t *testing.T) {
		_, err := g.Generate(0)
		if err == nil {
			t.Error("expected error for length 0, got nil")
		}

		_, err = g.Generate(-1)
		if err == nil {
			t.Error("expected error for length -1, got nil")
		}
	})
}
