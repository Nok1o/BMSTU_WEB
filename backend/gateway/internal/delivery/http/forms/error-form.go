package forms

//easyjson:json
type ErrorForm struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
