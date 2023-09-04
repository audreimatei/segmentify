package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	ErrorText      string `json:"detail"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(msg string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		ErrorText:      msg,
	}
}

func ErrNotFound(msg string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: http.StatusNotFound,
		ErrorText:      msg,
	}
}

func ErrRender(msg string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: http.StatusUnprocessableEntity,
		ErrorText:      msg,
	}
}

func ErrInternal(msg string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		ErrorText:      msg,
	}
}

func ValidationError(errs validator.ValidationErrors) *ErrResponse {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return &ErrResponse{
		HTTPStatusCode: http.StatusUnprocessableEntity,
		ErrorText:      strings.Join(errMsgs, ", "),
	}
}
