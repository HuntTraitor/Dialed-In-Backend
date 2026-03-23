package testutils

type Method struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ListMethodsResponse struct {
	Methods []Method `json:"methods"`
}
