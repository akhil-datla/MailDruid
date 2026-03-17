package handlers

// Response types for consistent API responses.

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type SummaryResponse struct {
	Summary string `json:"summary"`
	Image   string `json:"image,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func msgOK(msg string) *MessageResponse {
	return &MessageResponse{Message: msg}
}

func errResp(msg string) *ErrorResponse {
	return &ErrorResponse{Error: msg}
}
