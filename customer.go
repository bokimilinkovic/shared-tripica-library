package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"tripica-client/http"
	"tripica-client/http/errors"
)

const (
	customerBasePath = "/api/private/v1/agent/customer"

	customerPathGetByName = "/name/%s"
	customerPathGetByOUID = "/%s"
)

// Customer manages customer related endpoints within triPica.
type customerAPI struct {
	httpClient *http.Client
	address    string
}

// GetCustomerByOUID retrieves the customer using provided OUID.
func (c *customerAPI) GetCustomerByOUID(ouid string) (*Customer, error) {
	url := fmt.Sprintf(c.address+customerPathGetByOUID, ouid)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() == gohttp.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve customer with ouid %s: %w", ouid, err))
	}

	var customer Customer
	if err := json.Unmarshal(resp.Body(), &customer); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return &customer, nil
}

// GetCustomerByName retrieves a customer by the customerName <=> customer's external ID.
func (c *customerAPI) GetCustomerByName(customerName string) (*Customer, error) {
	url := fmt.Sprintf(c.address+customerPathGetByName, customerName)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() == gohttp.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve customer with customerName %s: %w", customerName, err))
	}

	var customer Customer
	if err := json.Unmarshal(resp.Body(), &customer); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return &customer, nil
}

// Customer represents a triPica customer.
type Customer struct {
	Name     string `json:"name"`
	OUID     string `json:"ouid"`
	PartyRef struct {
		PartyOUID string `json:"partyOuid"`
	} `json:"partyRef"`
	PaymentMeans []CustomerPaymentMean `json:"paymentMeans"`
}

// CustomerPaymentMean represents a triPica customer mean of payment.
type CustomerPaymentMean struct {
	EndDateTime     Date                               `json:"endDateTime"`
	Characteristics CustomerPaymentMeanCharacteristics `json:"characteristics"`
}

// CustomerPaymentMeanCharacteristics represents Characteristics part of CustomerPaymentMean.
type CustomerPaymentMeanCharacteristics struct {
	IBAN string `json:"iban"`
}
