package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"net/url"
	"strings"
	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/log"
)

const (
	productBasePath = "/api/private/v1/agent/product"

	productPathGetByCustomerOuid              = "/customerOuid/%s"
	productPathGetProductOrdersByCustomerOuid = "/productOrder/customerOuid/%s"
)

// Product manages product related endpoints within triPica.
type productAPI struct {
	httpClient *http.Client
	address    string

	logger log.Logger
}

// GetProductsByCustomerOUID retrieves products by UOID <=> unique internal identifier.
func (p *productAPI) GetProductsByCustomerOUID(customerOUID string, filter *ProductDateFilter) ([]Product, error) {
	url := fmt.Sprintf(p.address+productPathGetByCustomerOuid, customerOUID)

	if filter != nil {
		f, err := filter.Encode()
		if err != nil {
			return nil, err
		}

		url += "?filters=" + f
	}

	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() == gohttp.StatusNoContent {
		return []Product{}, nil
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve products with customerOUID %s: %s", customerOUID, err))
	}

	var products []Product
	if err := json.Unmarshal(resp.Body(), &products); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return products, nil
}

// GetProductOrdersByCustomerOUID retrieves product orders for a customer.
func (p *productAPI) GetProductOrdersByCustomerOUID(
	customerOUID string,
	filter *ProductDateFilter,
) ([]ProductOrder, error) {
	url := fmt.Sprintf(p.address+productPathGetProductOrdersByCustomerOuid, customerOUID)

	if filter != nil {
		f, err := filter.Encode()
		if err != nil {
			return nil, err
		}

		url += "?filters=" + f
	}

	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() == gohttp.StatusNoContent {
		return []ProductOrder{}, nil
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}

		return nil, NewTriPicaError(fmt.Errorf(
			"couldn't retrieve product orders with customerOUID %s: %w",
			customerOUID,
			err,
		))
	}

	var productOrders []ProductOrder
	if err := json.Unmarshal(resp.Body(), &productOrders); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return productOrders, nil
}

// Product represents a triPica product.
type Product struct {
	OUID                     string                 `json:"ouid"`
	Version                  int                    `json:"version"`
	DateTimeCreate           Date                   `json:"dateTimeCreate"`
	CreatorOUID              string                 `json:"creatorOuid"`
	DateTimeLastModif        Date                   `json:"dateTimeLastModif"`
	ModifierOUID             string                 `json:"modifierOuid"`
	StartDateTime            Date                   `json:"startDateTime"`
	EndDateTime              Date                   `json:"endDateTime"`
	Name                     string                 `json:"name"`
	OrderDate                Date                   `json:"orderDate"`
	NextRenewalDate          Date                   `json:"nextRenewalDate"`
	BillingAccountOuid       string                 `json:"billingAccountOuid"`
	ProductOfferingOuid      string                 `json:"productOfferingOuid"`
	ProductSpecificationOuid string                 `json:"productSpecificationOuid"`
	ProductSerialNumber      string                 `json:"productSerialNumber,omitempty"`
	RealizingService         string                 `json:"realizingService"`
	Status                   string                 `json:"status"`
	Characteristics          ProductCharacteristics `json:"characteristics"`
}

const (
	productSubscriptionPrefix = "SED4"
	// ActiveProductStatus is status of active product.
	ActiveProductStatus = "ACTIVE"
)

// IsSubscriptionProduct tells whether this product is a subscription product or not.
func (p Product) IsSubscriptionProduct() bool {
	return strings.HasPrefix(p.Name, productSubscriptionPrefix)
}

// ProductCharacteristics represents the Characteristics part of a product.
type ProductCharacteristics struct {
	MeterNumber string `json:"meterData.meterNumber"`
}

// IsActive returns true if product is in active status.
func (p *Product) IsActive() bool {
	return p.Status == ActiveProductStatus
}

// HasContract returns true if product has a contract.
func (p *Product) HasContract() bool {
	return p.ProductSerialNumber != ""
}

// ProductOrder represents a triPica product order.
type ProductOrder struct {
	OUID        string             `json:"ouid"`
	Category    string             `json:"category"`
	Description string             `json:"description"`
	OrderDate   Date               `json:"orderDate"`
	OrderItems  []ProductOrderItem `json:"orderItems"`
}

// ProductOrderItem represents a triPica product order item.
type ProductOrderItem struct {
	Product            Product `json:"product"`
	BillingAccountOUID string  `json:"billingAccountOuid"`
}

// ProductDateFilter is used to filter products by their endDateTime and startDateTime values.
type ProductDateFilter struct {
	Begin *int `json:"begin,omitempty"` // Maximum product endDateTime in milliseconds.
	End   *int `json:"end,omitempty"`   // Minimum product startDateTime in milliseconds.
}

// Encode encodes the filter, allowing it to be used when making HTTP requests to triPica.
func (f *ProductDateFilter) Encode() (string, error) {
	v, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	return url.QueryEscape(string(v)), nil
}

// ProductsAssociatedToBillingAccounts  filters products based on their association with provided billing accounts.
// In order to get products relevant only for certain MBA and its CBAs, filter them first separately.
func ProductsAssociatedToBillingAccounts(products []Product, billingAccounts []*BillingAccount) []Product {
	relevantProducts := products[:0]

	for _, p := range products {
		for _, a := range billingAccounts {
			if p.BillingAccountOuid == a.OUID {
				relevantProducts = append(relevantProducts, p)
			}
		}
	}

	return relevantProducts
}

// SubscriptionProducts filters products based on whether they're a subscription product or not.
func SubscriptionProducts(products []Product) []Product {
	subscriptionProducts := products[:0]

	for _, p := range products {
		if p.IsSubscriptionProduct() {
			subscriptionProducts = append(subscriptionProducts, p)
		}
	}

	return subscriptionProducts
}
