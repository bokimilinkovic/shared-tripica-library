package http

import (
	"net/http"
)

// Response is a wrapper for the received HTTP response.
type Response struct {
	body        []byte
	rawResponse *http.Response
}

// NewResponse returns a new Response.
func NewResponse(body []byte, rawResponse *http.Response) *Response {
	return &Response{
		body:        body,
		rawResponse: rawResponse,
	}
}

// Body returns the response body as a []byte array.
func (r *Response) Body() []byte {
	return r.body
}

// Header returns the response header.
func (r *Response) Header() *http.Header {
	return &r.rawResponse.Header
}

// Cookies retruns all the response cookies.
func (r *Response) Cookies() []*http.Cookie {
	return r.rawResponse.Cookies()
}

// StatusCode retruns the response status code.
func (r *Response) StatusCode() int {
	return r.rawResponse.StatusCode
}
