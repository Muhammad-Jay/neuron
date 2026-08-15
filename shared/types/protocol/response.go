package protocol

// Response Base wrapper to standardize all API replies
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}
