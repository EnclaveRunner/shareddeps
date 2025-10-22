package api

import "context"

// ensure that we've conformed to the `ServerInterface` with
// a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

// GetHealth implements StrictServerInterface.
func (s *Server) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200Response{}, nil
}
