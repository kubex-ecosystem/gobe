package types

// ErrorResponse defines the error payload returned by the API.
type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"detailed error message"`
}

// StatusResponse represents a generic status acknowledgement.
type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

// MessageResponse describes a response that returns a status with an informational message.
type MessageResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"operation completed"`
}
