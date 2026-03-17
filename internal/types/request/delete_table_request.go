package request

type DeleteTableRequest struct {
	Tables []string `json:"tables" validate:"required,min=1"`
}
