package tripica

import (
	"fmt"
	"sync"
	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/jwt"
	"tripica-client/log"
)

// Client allows HTTP communication with a triPica server.
// Every tripica separated module(biling,customer,individual...) are built in this client.
type Client struct {
	address     string
	credentials Credentials
	token       *jwt.Token
	mux         sync.Mutex
	httpClient  *http.Client
	logger      log.Logger

	*loginAPI
	*billingAPI
	*customerAPI
	*individualAPI
	*productAPI
	*networkEntityAPI
	*notifyAPI
}

// Config configures the required information for accessing triPica endpoints.
type Config struct {
	Host        string
	Credentials Credentials
}

// Credentials objects hold data allowing the service to be authenticated by triPica.
// While password is required, only one of the remaining attributes is necessary.
// If both email and alias are submitted, they both need to be correct.
type Credentials struct {
	Email    string
	Password string
	Alias    string
}

// NewClient returns Client for communication to tripica.
func NewClient(config Config, client *http.Client, logger log.Logger) *Client {
	c := &Client{
		address:     config.Host,
		credentials: config.Credentials,
		logger:      logger,
	}

	tokenHolder := &Client{
		credentials: config.Credentials,
	}

	client.Apply(
		http.WithAuthToken(tokenHolder),
	)

	c.httpClient = client

	c.loginAPI = &loginAPI{
		httpClient:      client,
		address:         c.address,
		addressAgent:    c.address + loginBasePathAgent,
		addressCustomer: c.address + loginBasePathCustomer,
		logger:          logger,
	}

	c.billingAPI = &billingAPI{
		httpClient: client,
		address:    c.address + billingBasePath,
		logger:     logger,
	}

	c.customerAPI = &customerAPI{
		httpClient: client,
		address:    c.address + customerBasePath,
	}

	c.individualAPI = &individualAPI{
		httpClient: client,
		address:    c.address + individualBasePath,
		logger:     logger,
	}

	c.networkEntityAPI = &networkEntityAPI{
		httpClient: client,
		address:    c.address + networkEntityBasePath,
		logger:     logger,
	}

	c.productAPI = &productAPI{
		httpClient: client,
		address:    c.address + productBasePath,
		logger:     logger,
	}

	c.notifyAPI = &notifyAPI{
		httpClient: client,
		address:    c.address,
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
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.token != nil && !c.token.IsExpired() {
		return nil
	}

	c.token = nil

	token, err := c.loginAPI.authorize(c.credentials)
	if err != nil {
		authErr := &errors.AuthorizationError{Err: err}
		return fmt.Errorf("couldn't authorize with triPica: %s", authErr)
	}

	c.token = token

	return nil
}

// RawToken returns the raw underlying token.
func (c *Client) RawToken() string {
	return c.token.Raw
}
