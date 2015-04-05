package kiri

type Service struct {
	Name    string                 `json:"name"`
	Address string                 `json:"address"`
	Tags    map[string]interface{} `json:"tags,omitempty"`
}
