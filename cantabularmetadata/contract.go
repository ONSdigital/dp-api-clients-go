package cantabularmetadata

// ErrorResponse models the error response from cantabular
type ErrorResponse struct {
	Message string `json:"message"`
}

type GetDefaultClassificationRequest struct {
	Dataset   string
	Variables []string
}

type GetDefaultClassificationResponse struct {
	Variables []string
}

type Data struct {
	Dataset `json:"dataset"`
}

type Dataset struct {
	Vars []Var `json:"vars"`
}

type Var struct {
	Name string `json:"name"`
	Meta Meta   `json:"meta"`
}

type Meta struct {
	DefaultClassificationFlag string `json:"Default_Classification_Flag"`
}
