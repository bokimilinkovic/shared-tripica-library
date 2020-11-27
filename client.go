package tripica

import (
	"shared-tripica-library/http"
	"shared-tripica-library/jwt"
	"shared-tripica-library/log"
	"sync"
)

// Client allows HTTP communication with a triPica server.
// Every tripica separated module(biling,customer,individual...) are built in this client
type Client struct {
	address     string
	credentials Credentials
	token       *jwt.Token
	mux         sync.Mutex
	httpClient  *http.Client
	logger      log.Logger

	*billingAPI
}

// Config configures the required information for accessing triPica endpoints.
type Config struct {
	Host        string
	Credentials Credentials
}

// Credentials objects hold data allowing the service to be authenticated by triPica.
type Credentials struct {
	Email    string
	Password string
}

// NewClient returns Client for communication to tripica.
func NewClient(config Config, client *http.Client, logger log.Logger) *Client {
	c := &Client{
		address:     config.Host,
		credentials: config.Credentials,
		logger:      logger,
	}

	client.Apply(
		http.WithAuthToken(c),
		http.JSONClient(),
	)

	c.httpClient = client

	c.billingAPI = &billingAPI{
		httpClient: client,
		address:    c.address + billingBasePath,
		logger:     logger,
	}

	return c
}

// InvalidateToken sets the authorization token to nil.
func (c *Client) InvalidateToken() {
	c.mux.Lock()
	c.token = nil
	c.mux.Unlock()
}

// RefreshToken checks whether the token is valid and fetches a new one if it isn't.
func (c *Client) RefreshToken() error {
	// To be implemented...
	return nil
}

// RawToken returns the raw underlying token.
func (c *Client) RawToken() string {
	return c.token.Raw
}
