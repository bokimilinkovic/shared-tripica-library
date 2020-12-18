package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"time"
	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/log"
)

const (
	individualBasePath = "/api/private/v1/agent/individual"

	individualPathGetByPartyOUID = "/%s"
)

// Individual manages individual related endpoints within triPica.
type individualAPI struct {
	httpClient *http.Client
	address    string

	logger log.Logger
}

// GetIndividualByPartyOUID retrieves an individual by the customer's party OUID.
func (i *individualAPI) GetIndividualByPartyOUID(partyOUID string) (*Individual, error) {
	url := fmt.Sprintf(i.address+individualPathGetByPartyOUID, partyOUID)

	resp, err := i.httpClient.Get(url)
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

		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve individual with partyOUID %s: %w", partyOUID, err))
	}

	var individual *Individual
	if err := json.Unmarshal(resp.Body(), &individual); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return individual, nil
}

const (
	contactMediumTypeDeliveryAddress = ContactMediumType("DELIVERY_ADDRESS")
	contactMediumTypeBillingAddress  = ContactMediumType("BILLING_ADDRESS")
)

// Individual represents a triPica individual.
type Individual struct {
	OUID           string          `json:"ouid"`
	Title          string          `json:"title"`
	Name           string          `json:"givenName"`
	LastName       string          `json:"familyName"`
	Gender         string          `json:"gender"`
	ContactMediums []ContactMedium `json:"contactMediums"`
}

// DeliveryAddress returns the delivery address medium.
func (i *Individual) DeliveryAddress(now time.Time) MediumTypeAddress {
	return i.contactMedium(now, contactMediumTypeDeliveryAddress).Medium.MediumTypeAddress
}

// BillingAddress returns the billing address medium.
func (i *Individual) BillingAddress(now time.Time) MediumTypeAddress {
	return i.contactMedium(now, contactMediumTypeBillingAddress).Medium.MediumTypeAddress
}

// contactMedium searches through the contact mediums for the specified type.
// It returns an empty contact medium in case none is found, since zero values
// will be omitted when sending them to collectAI.
func (i *Individual) contactMedium(now time.Time, typ ContactMediumType) ContactMedium {
	for _, cm := range i.ContactMediums {
		if cm.Type == typ &&
			// We should only consider mediums which have startDateTime in the past,
			// and EndDateTime which is either nil or in the future.
			cm.StartDateTime.Before(now) && (cm.EndDateTime == nil || now.Before(cm.EndDateTime.Time)) {
			return cm
		}
	}

	return ContactMedium{}
}

// ContactMediumType represents a possible contact media type.
type ContactMediumType string

// ContactMedium represents triPica individual's contact mediums.
type ContactMedium struct {
	OUID          string            `json:"ouid"`
	Preferred     bool              `json:"preferred"`
	Type          ContactMediumType `json:"type"`
	StartDateTime Date              `json:"startDateTime"`
	EndDateTime   *Date             `json:"endDateTime"`
	Medium        `json:"medium"`
}

// Medium represents a single Medium.
type Medium struct {
	OUID string `json:"ouid"` //nolint: misspell
	MediumTypeAddress
	Type string `json:"type"`
}

// MediumTypeAddress represents one of the possible medium types for an individual.
type MediumTypeAddress struct {
	City     string `json:"city"`
	Country  string `json:"country"`
	Street1  string `json:"street1"`
	Street2  string `json:"street2"`
	Postcode string `json:"postCode"`
}
