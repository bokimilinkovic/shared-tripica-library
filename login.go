package tripica

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/jwt"
	"tripica-client/log"
)

const (
	loginBasePathPrivateCustomer = "/api/private/v1/login"
	loginBasePathAgent    = "/api/private/v1/agent/login"
	loginBasePathCustomer = "/api/v1/login"
	loginPathGetByCustomerOUID = "/customerOuid/%s"
	loginPathGenerateJWT       = "/jwt"
)

// loginAPI manages login related endpoints within triPica.
type loginAPI struct {
	httpClient      *http.Client
	address 		string
	addressAgent    string
	addressCustomer string
	logger 			log.Logger
}

// GetLoginByCustomerOUID retrieves login info using customer OUID.
func (l *loginAPI) GetLoginByCustomerOUID(customerOUID string) (*Login, error) {
	url := fmt.Sprintf(l.addressAgent+loginPathGetByCustomerOUID, customerOUID)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() == stdhttp.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode() != stdhttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve login with customerOUID %s: %w", customerOUID, err))
	}

	var multipleLoginResp MultipleLoginResponse
	if err := json.Unmarshal(resp.Body(), &multipleLoginResp); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	if len(multipleLoginResp) == 0 {
		return nil, nil
	}

	return &multipleLoginResp[0], nil
}

func (l *loginAPI) GetLoginInfoForToken(token string) (*Login, error) {
	url := fmt.Sprintf(l.address + loginBasePathPrivateCustomer)

	resp, err := l.httpClient.Get(
		url,
		http.SkipAuthToken(),
		http.WithUserToken(token),
		http.InvalidateCookie("trpcCookie"),
	)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != stdhttp.StatusOK {
		return nil, errors.NewHTTPError(nil, resp.Body(), resp.StatusCode())
	}

	login := &Login{}
	if err := json.Unmarshal(resp.Body(), &login); err != nil {
		return nil, errors.NewParseError(err, resp.Body())
	}

	return login, nil
}

func (l *loginAPI) authorize(creds Credentials) (*jwt.Token, error) {
	url := l.addressCustomer + loginPathGenerateJWT

	reqBody := NewTokenRequest(creds.Email, creds.Alias, creds.Password)
	resp, err := l.httpClient.Post(url, reqBody, http.SkipAuthToken())
	
	if err != nil {
		return nil, errors.NewHTTPRequestError(err)
	}

	if resp.StatusCode() != stdhttp.StatusCreated {
		return nil, errors.NewHTTPError(nil, resp.Body(), resp.StatusCode())
	}

	tokenResponse := &TokenResponse{}
	if err := json.Unmarshal(resp.Body(), &tokenResponse); err != nil {
		return nil, errors.NewParseError(err, resp.Body())
	}

	token, err := jwt.NewToken(tokenResponse.Token)
	if err != nil {
		return nil, fmt.Errorf("couldn't create new token: %w", err)
	}

	return token, nil
}

// MultipleLoginResponse represents triPica response that contains multiple Login objects.
type MultipleLoginResponse []Login

// Login represents a triPica login.
type Login struct {
	Email string `json:"email"`
}

// TokenRequest represents a request for a login token within triPica.
type TokenRequest struct {
	Email      string `json:"email,omitempty"`
	Alias      string `json:"alias,omitempty"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// TokenResponse represents a response to a token request.
type TokenResponse struct {
	Token string `json:"token"`
}

// NewTokenRequest creates a new token request.
func NewTokenRequest(email, alias, password string) *TokenRequest {
	return &TokenRequest{
		Email:      email,
		Alias:      alias,
		Password:   password,
		RememberMe: true,
	}
}
