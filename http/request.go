package http

import (
	resty "github.com/go-resty/resty/v2"
	"net/http"
)

type request struct {
	client        *Client
	baseRequest   *resty.Request
	url           string
	method        string
	body          interface{}
	shouldRepeat  bool
	skipAuthToken bool
}

// RequestOption represents a functional option used to initialize a Reqeust.
type RequestOption func(*request) *request

func (c *Client) newRequest(url, method string, body interface{}, options ...RequestOption) *request {
	r := &request{
		url:          url,
		method:       method,
		body:         body,
		shouldRepeat: true,
		baseRequest:  c.retryer.NewRequest(),
		client:       c,
	}

	if r.body != nil {
		r.baseRequest.SetBody(body)
	}

	for _, option := range options {
		r = option(r)
	}

	return r
}

func (r *request) setAuthToken(token string) {
	r.baseRequest.SetAuthToken(token)
}

// SkipAuthToken disables the WithAuthToken middleware for the request.
func SkipAuthToken() RequestOption {
	return func(r *request) *request {
		r.skipAuthToken = true

		return r
	}
}

// QueryParams sets query params on the request.
func QueryParams(params map[string]string) RequestOption {
	return func(r *request) *request {
		r.baseRequest.SetQueryParams(params)

		return r
	}
}

// PostForm sets the Content-Type header to form.
func PostForm() RequestOption {
	return func(r *request) *request {
		r.baseRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		return r
	}
}

// JSONContent sets the Content-Type header to application/json.
func JSONContent() RequestOption {
	return func(r *request) *request {
		r.baseRequest.Header.Add("Content-Type", "application/json")

		return r
	}
}

func WithUserToken(token string) RequestOption {
	return func(r *request) *request {
		r.setAuthToken(token)

		return r
	}
}

func InvalidateCookie(cookieName string) RequestOption {
	return func(r *request) *request {
		r.baseRequest.SetCookie(&http.Cookie{
			Name:  cookieName,
			Value: "",
		})

		return r
	}
}

func (r *request) execute() (*Response, error) {
	for _, b := range r.client.beforeRequest {
		if err := b(r); err != nil {
			return nil, err
		}
	}

	resp, err := r.baseExecute()
	if err != nil {
		return nil, err
	}

	for _, a := range r.client.afterRequest {
		if resp, err = a(resp, r); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (r *request) baseExecute() (*Response, error) {
	baseResponse, err := r.baseRequest.Execute(r.method, r.url)
	if err != nil {
		return nil, err
	}

	return NewResponse(baseResponse.Body(), baseResponse.RawResponse), nil
}
