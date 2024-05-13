package dto

type (
	StateResponse struct {
		NodeSize      int64 `json:"node_size"`
		NodeAvailable int64 `json:"node_available"`
		NodeUsed      int64 `json:"node_used"`
	}
)
