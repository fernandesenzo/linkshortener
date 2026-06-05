package handler

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	Code string `json:"code"`
}
