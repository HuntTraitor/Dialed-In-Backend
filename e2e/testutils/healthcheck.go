package testutils

type Response struct {
	Status     string `json:"status"`
	SystemInfo struct {
		Environment string `json:"environment"`
		Version     string `json:"version"`
	} `json:"system_info"`
}
