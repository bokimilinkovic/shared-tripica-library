package http

import (
	"net/http"
	"net/http/httptest"
	"shared-tripica-library/log"
	"testing"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type (
	handler struct{ clientID string }
)

// ServeHTTP handles requests sent to handler. It reads and stores the client_id header value.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.clientID = r.Header.Get("client_id")
}

// TestClient verifies that the client configuration is applied to every HTTP request.
func TestClient(t *testing.T) {
	assert := assert.New(t)

	h := handler{}
	srv := httptest.NewServer(&h)
	defer srv.Close()

	client := &Client{retryer: resty.New()}
	clientID := "restytest"
	client.retryer.SetHeader("client_id", clientID)

	t.Run("Get method uses the struct's client value to send a request", func(t *testing.T) {
		_, err := client.Get(srv.URL)
		assert.NoError(err)
		assert.Equal(clientID, h.clientID)
	})

	t.Run("Delete method uses the struct's client value to send a request", func(t *testing.T) {
		res, err := client.Delete(srv.URL, nil)
		assert.NoError(err)
		assert.NotNil(res)
		assert.Equal(clientID, h.clientID)
	})

	t.Run("Patch method uses the struct's client value to send a request", func(t *testing.T) {
		_, err := client.Patch(srv.URL, nil)
		assert.NoError(err)
		assert.Equal(clientID, h.clientID)
	})

	t.Run("Post method uses the struct's client value to send a request", func(t *testing.T) {
		_, err := client.Post(srv.URL, nil)
		assert.NoError(err)
		assert.Equal(clientID, h.clientID)
	})

	t.Run("Put method uses the struct's client value to send a request", func(t *testing.T) {
		_, err := client.Put(srv.URL, nil)
		assert.NoError(err)
		assert.Equal(clientID, h.clientID)
	})
}

// TestNewClient verifies that the configuration is properly applied to the client.
func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(log.NewTestLogger(), ConfigureRetryer(NewRetryerConfig(2, 3, 4, 5)))
	retryer := client.retryer

	assert.Equal(2, retryer.RetryCount)
	assert.Equal(time.Duration(3000000), retryer.RetryWaitTime)
	assert.Equal(time.Duration(4000000), retryer.RetryMaxWaitTime)
	assert.Equal(time.Duration(5000000), retryer.GetClient().Timeout)
	assert.Len(retryer.RetryConditions, 1)
}

// TestNewClient verifies that the configuration is properly applied to the client.
func TestDefaultClient(t *testing.T) {
	assert := assert.New(t)
	client := DefaultClient(log.NewTestLogger())
	retryer := client.retryer

	assert.Equal(4, retryer.RetryCount)
	assert.Equal(time.Duration(1000000000), retryer.RetryWaitTime)
	assert.Equal(time.Duration(2000000000), retryer.RetryMaxWaitTime)
	assert.Equal(time.Duration(5000000000), retryer.GetClient().Timeout)
	assert.Len(retryer.RetryConditions, 1)
}
