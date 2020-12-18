package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"

	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/log"
)

const (
	networkEntityBasePath = "/api/private/v1/agent/networkEntity"

	networkEntityPathGetNetworkEntityBySubscriptionOuid = "/subscriptionOuid/%s"
)

type networkEntityAPI struct {
	httpClient *http.Client
	address    string

	logger log.Logger
}

// GetNetworkEntityForSubscriptionOuid returns a NetworkEntity by subscriptionOuid.
func (e *networkEntityAPI) GetNetworkEntityBySubscriptionOuid(subscriptionOuid string) (*NetworkEntity, error) {
	url := fmt.Sprintf(e.address+networkEntityPathGetNetworkEntityBySubscriptionOuid, subscriptionOuid)

	resp, err := e.httpClient.Get(url)
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

		return nil, NewTriPicaError(fmt.Errorf(
			"couldn't retrieve network entity with subscription ouid %s: %w",
			subscriptionOuid,
			err,
		))
	}

	var networkEntity *NetworkEntity
	if err := json.Unmarshal(resp.Body(), &networkEntity); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return networkEntity, nil
}

// GetMeterNumbersForProducts retrieves meter numbers for the provided products. It is assumed
// product filtering was already done, and list of prodived products only contain subscription products.
func (e *networkEntityAPI) GetMeterNumbersForProducts(products []Product) ([]string, error) {
	networkEntities := []*NetworkEntity{}

	for _, p := range products {
		networkEntity, err := e.GetNetworkEntityBySubscriptionOuid(p.OUID)
		if err != nil {
			return nil, err
		}

		networkEntities = append(networkEntities, networkEntity)
	}

	meterNumbers := []string{}

	for _, e := range networkEntities {
		for _, i := range e.NetworkItems {
			if i.Characteristics.MeterNumber != "" {
				meterNumbers = append(meterNumbers, i.Characteristics.MeterNumber)
			}
		}
	}

	return meterNumbers, nil
}

// NetworkEntityDTO represents a triPica network entity.
type NetworkEntity struct {
	OUID             string              `json:"ouid"`
	Type             string              `json:"type"`
	SubscriptionOUID string              `json:"subscriptionOuid"`
	NetworkItems     []NetworkEntityItem `json:"networkItem"`
}

// NetworkEntityItem represents a triPica network entity item.
type NetworkEntityItem struct {
	Characteristics NetworkEntityItemCharacteristics `json:"characteristics"`
}

// NetworkEntityItemCharacteristics contains characterics part of NetworkEntityItem.
type NetworkEntityItemCharacteristics struct {
	MeterNumber string `json:"meterData.meterNumber"`
}
