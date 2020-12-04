package http_test

import (
	"errors"
	"io/ioutil"
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"tripica-client/http"
	httpmock "tripica-client/http/mock"

	"tripica-client/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	handler struct {
		request []byte
		query   url.Values
		header  stdhttp.Header
		require *require.Assertions
	}
	unauthorizedHandler struct{}
	badRequestHandler   struct {
		require *require.Assertions
	}
)

var methods = []string{stdhttp.MethodGet, stdhttp.MethodPatch, stdhttp.MethodPost, stdhttp.MethodPut, stdhttp.MethodDelete} //nolint
// const log = logrus.New()

// ServeHTTP handles requests to the handler. It reads the authorization token from request header
// and stores it in the token field.
func (h *handler) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	h.query = r.URL.Query()
	h.header = r.Header
	if r.Method != stdhttp.MethodGet {
		h.readBody(w, r)
	}
	w.WriteHeader(stdhttp.StatusOK)
	_, err := w.Write([]byte("message"))
	h.require.NoError(err)
}

func (h *handler) readBody(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	h.request = body
}

func (h *handler) token() string {
	return h.header.Get("Authorization")
}

// ServeHTTP handles requests to badRequestHandler. It responds with StatusBadRequest.
func (h *badRequestHandler) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.WriteHeader(stdhttp.StatusBadRequest)
	_, err := w.Write([]byte("bad request"))
	h.require.NoError(err)
}

// ServeHTTP handles requests to unauthorizedHandler. It responds with statusUnauthorized.
func (h *unauthorizedHandler) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	w.WriteHeader(stdhttp.StatusUnauthorized)
}

func TestClient_Get(t *testing.T) {
	runMethodTests(stdhttp.MethodGet, t)
}

func TestClient_Post(t *testing.T) {
	runMethodTests(stdhttp.MethodPost, t)
}

func TestClient_Patch(t *testing.T) {
	runMethodTests(stdhttp.MethodPatch, t)
}

func TestClient_Put(t *testing.T) {
	runMethodTests(stdhttp.MethodPut, t)
}

func TestClient_Delete(t *testing.T) {
	runMethodTests(stdhttp.MethodDelete, t)
}

func TestWithAuthToken(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for _, method := range methods {
		method := method
		t.Run(method+" request with auth token is successfully sent", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			tokenHolder := &httpmock.TokenHolder{}
			tokenHolder.On("InvalidateToken").Return()
			tokenHolder.On("RefreshToken").Return(nil)
			tokenHolder.On("RawToken").Return("token")
			client := http.NewClient(log.NewTestLogger(), http.WithAuthToken(tokenHolder))

			res, err := executeRequest(client, srv.URL, method, nil)
			assert.NoError(err)

			assert.Equal(stdhttp.StatusOK, res.StatusCode())
			assert.Equal("Bearer token", h.token())
		})

		t.Run(method+" request with auth token is repeated if the server responds with Unauthorized", func(t *testing.T) {
			h := unauthorizedHandler{}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			tokenHolder := &httpmock.TokenHolder{}
			tokenHolder.On("RefreshToken").Return(nil).Twice()
			tokenHolder.On("InvalidateToken").Return().Twice()
			tokenHolder.On("RawToken").Return("token").Twice()
			client := http.NewClient(log.NewTestLogger(), http.WithAuthToken(tokenHolder))

			res, err := executeRequest(client, srv.URL, method, nil)
			assert.NoError(err)
			assert.Equal(res.StatusCode(), stdhttp.StatusUnauthorized)
			tokenHolder.AssertExpectations(t)
		})

		t.Run(method+" request fails due to error occurring while refreshing auth token", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			RefreshError := errors.New("refresh token error")
			tokenHolder := &httpmock.TokenHolder{}
			tokenHolder.On("InvalidateToken").Return()
			tokenHolder.On("RefreshToken").Return(RefreshError)
			tokenHolder.On("RawToken").Return(nil)

			client := http.NewClient(log.NewTestLogger(), http.WithAuthToken(tokenHolder))

			res, err := executeRequest(client, srv.URL, method, nil)
			assert.EqualError(err, RefreshError.Error())
			assert.Nil(res)
		})

		t.Run(method+" request skips setting auth token when configured so", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			tokenHolder := &httpmock.TokenHolder{}
			client := http.NewClient(log.NewTestLogger(), http.WithAuthToken(tokenHolder))

			res, err := executeRequest(client, srv.URL, method, nil, http.SkipAuthToken())
			assert.NoError(err)
			assert.Equal(stdhttp.StatusOK, res.StatusCode())
			assert.Empty(h.token())
		})
	}
}

func TestQueryParams(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	for _, method := range methods {
		method := method
		t.Run(method+" request with query params is successfully sent", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			client := http.NewClient(log.NewTestLogger())

			params := http.QueryParams(map[string]string{
				"test": "params",
			})
			res, err := executeRequest(client, srv.URL, method, nil, params)
			assert.NoError(err)

			assert.Equal(stdhttp.StatusOK, res.StatusCode())
			assert.Contains(h.query, "test")
		})
	}
}

func TestJSONClient(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	for _, method := range methods {
		method := method
		t.Run(method+" request with json headers is successfully sent", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			client := http.NewClient(log.NewTestLogger(), http.JSONClient())

			res, err := executeRequest(client, srv.URL, method, nil)
			assert.NoError(err)

			assert.Equal(stdhttp.StatusOK, res.StatusCode())
			assert.Equal(h.header.Get("Accept"), "application/json")
			assert.Equal(h.header.Get("Content-Type"), "application/json")
		})
	}
}

func TestApply(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	for _, method := range methods {
		method := method
		t.Run(method+" request is successfully sent with the applied options", func(t *testing.T) {
			h := handler{require: require}
			srv := httptest.NewServer(&h)
			defer srv.Close()

			client := http.NewClient(log.NewTestLogger())
			client.Apply(http.JSONClient())
			res, err := executeRequest(client, srv.URL, method, nil)
			assert.NoError(err)

			assert.Equal(stdhttp.StatusOK, res.StatusCode())
			assert.Equal(h.header.Get("Accept"), "application/json")
			assert.Equal(h.header.Get("Content-Type"), "application/json")
		})
	}
}

func runMethodTests(method string, t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	body := "body"

	t.Run(method+" request is successfully sent", func(t *testing.T) {
		h := handler{require: require}
		srv := httptest.NewServer(&h)
		defer srv.Close()

		client := http.NewClient(log.NewTestLogger())

		res, err := executeRequest(client, srv.URL, method, body)
		assert.NoError(err)

		if method != stdhttp.MethodGet {
			assert.Equal(h.request, []byte(body))
		}
		assert.Equal(stdhttp.StatusOK, res.StatusCode())
		assert.Equal([]byte("message"), res.Body())
	})

	t.Run(method+" request without body is successfully sent", func(t *testing.T) {
		h := handler{require: require}
		srv := httptest.NewServer(&h)
		defer srv.Close()

		client := http.NewClient(log.NewTestLogger())

		res, err := executeRequest(client, srv.URL, method, nil)
		assert.NoError(err)

		if method != stdhttp.MethodGet {
			assert.Equal([]byte{}, h.request)
		}
		assert.Equal(stdhttp.StatusOK, res.StatusCode())
		assert.Equal([]byte("message"), res.Body())
	})

	t.Run(method+" request to invalid URL fails", func(t *testing.T) {
		client := http.NewClient(log.NewTestLogger())

		res, err := executeRequest(client, "invalidurl", method, nil)
		assert.Error(err)
		assert.Nil(res)
	})

	t.Run(method+" request returns unexpected status code", func(t *testing.T) {
		h := badRequestHandler{require: require}
		srv := httptest.NewServer(&h)
		defer srv.Close()

		client := http.NewClient(log.NewTestLogger())

		res, err := executeRequest(client, srv.URL, method, nil)
		assert.NoError(err)
		assert.Equal(res.StatusCode(), stdhttp.StatusBadRequest)
	})

	t.Run(method+" request returns unauthorized status code", func(t *testing.T) {
		h := unauthorizedHandler{}
		srv := httptest.NewServer(&h)
		defer srv.Close()

		client := http.NewClient(log.NewTestLogger())

		res, err := executeRequest(client, srv.URL, method, nil)
		assert.NoError(err)
		assert.Equal(res.StatusCode(), stdhttp.StatusUnauthorized)
	})
}

func executeRequest(
	client *http.Client,
	url, method string,
	body interface{},
	options ...http.RequestOption,
) (*http.Response, error) {
	var res *http.Response
	var err error
	switch method {
	case stdhttp.MethodGet:
		res, err = client.Get(url, options...)
	case stdhttp.MethodPatch:
		res, err = client.Patch(url, body, options...)
	case stdhttp.MethodPost:
		res, err = client.Post(url, body, options...)
	case stdhttp.MethodPut:
		res, err = client.Put(url, body, options...)
	case stdhttp.MethodDelete:
		res, err = client.Delete(url, body, options...)
	}

	return res, err
}
