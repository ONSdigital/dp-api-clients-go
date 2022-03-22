package files

type jsonError struct {
	Code        string `json:"errorCode"`
	Description string `json:"description"`
}

type jsonErrors struct {
	Errors []jsonError `json:"errors"`
}
