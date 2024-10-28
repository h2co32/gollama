package middleware

import "net/http"

// Middleware interface for request/response lifecycle handling
type Middleware interface {
	ProcessRequest(req *http.Request) (*http.Request, error)
	ProcessResponse(res *http.Response) (*http.Response, error)
}
