package mock

import (
	"github.com/stretchr/testify/mock"
	"tripica-client/http"
)

// Client mocks an http.Client object.
type Client struct {
	mock.Mock
}

func (c *Client) Apply(options ...http.ClientOption) {
	c.Called(options)
}

func (c *Client) Get(url string, options ...http.RequestOption) (*http.Response, error) {
	args := c.Called(url, options)
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	return nil, args.Error(1)
}

func (c *Client) Post(url string, body interface{}, options ...http.RequestOption) (*http.Response, error) {
	args := c.Called(url, body, options)
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	return nil, args.Error(1)
}

func (c *Client) Patch(url string, body interface{}, options ...http.RequestOption) (*http.Response, error) {
	args := c.Called(url, body, options)
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	return nil, args.Error(1)
}

func (c *Client) Put(url string, body interface{}, options ...http.RequestOption) (*http.Response, error) {
	args := c.Called(url, body, options)
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	return nil, args.Error(1)
}

func (c *Client) Delete(url string, body interface{}, options ...http.RequestOption) (*http.Response, error) {
	args := c.Called(url, body, options)
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}

	return nil, args.Error(1)
}
