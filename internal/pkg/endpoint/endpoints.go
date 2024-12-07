package endpoint

import "context"

// An Endpoint represents a single operation. Typical endpoint function implementations adapt a request
// to a service call, and adapt the response from the service call to a response and error.
type Endpoint func(ctx context.Context, request any) (any, error)
