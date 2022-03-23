package app

import "strconv"

type ExampleRequest struct {
	Message string `json:"message"`
}

type ExampleResponse struct {
	Message string `json:"message"`
}

func (a *Application) Example(msg ExampleRequest) (*ExampleResponse, error) {
	a.responseCounter++

	return &ExampleResponse{
		Message: "Відповідь " + strconv.Itoa(a.responseCounter) + ": " + msg.Message}, nil
}
