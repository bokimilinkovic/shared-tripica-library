package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"tripica-client/http"
	"tripica-client/http/errors"
	"tripica-client/log"
	"strings"
)

const (
	billingBasePath = "/api/private/v1/agent/billing"

	billingPathGetDueBillingAccountBalancesByCustomer = "/billingAccountBalance/customerOuid/%s/status/DUE"
	billingPathGetAppliedBillingCharges               = "/appliedBillingCharge?filters=transactionIds="
	billingPathGetListOfSettlementNodeAdviceByAccount = "/settlement/billingAccountOuid/%s"
	billingPathGetBillingAccountByMBA                 = "/billingAccount/name/%s"
	billingPathGetBillingAccountsByCustomer           = "/billingAccount/customerOuid/%s"
)

// Billing manages billing related endpoints within triPica.
type billingAPI struct {
	httpClient *http.Client
	address    string

	logger log.Logger
}

// GetBillingAccountByMBA retrieves a billing account using provided MBA.
func (b *billingAPI) GetBillingAccountByMBA(mba string) (*BillingAccount, error) {
	url := fmt.Sprintf(b.address+billingPathGetBillingAccountByMBA, mba)

	resp, err := b.httpClient.Get(url)
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
		return nil, NewTriPicaError(fmt.Errorf("couldn't retrieve billing account with mba %s: %w", mba, err))
	}

	var billingAccount *BillingAccount
	if err := json.Unmarshal(resp.Body(), &billingAccount); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return billingAccount, nil
}

// GetDueBillingAccountBalancesByCustomer retrieves account balances for the customer that are due.
func (b *billingAPI) GetDueBillingAccountBalancesByCustomer(customerOUID string) ([]*BillingAccountBalance, error) {
	url := fmt.Sprintf(b.address+billingPathGetDueBillingAccountBalancesByCustomer, customerOUID)

	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}
		return nil, NewTriPicaError(
			fmt.Errorf("couldn't retrieve billing balances with customerOUID %s: %w", customerOUID, err),
		)
	}

	var accountBalances []*BillingAccountBalance
	if err := json.Unmarshal(resp.Body(), &accountBalances); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return accountBalances, nil
}

// GetAppliedBillingChargesByTransactionIDs retrieves applied billing charges related to the transaction IDs.
func (b *billingAPI) GetAppliedBillingChargesByTransactionIDs(transactionIDs string) ([]*AppliedBillingCharge, error) {
	url := b.address + billingPathGetAppliedBillingCharges + transactionIDs

	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}
		return nil, NewTriPicaError(
			fmt.Errorf("couldn't retrieve billing charges with transactionIDs %s: %w", transactionIDs, err),
		)
	}

	var charges []*AppliedBillingCharge
	if err := json.Unmarshal(resp.Body(), &charges); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return charges, nil
}

// GetSettlementNoteAdviceByBillingAccount retrieves settlement note advices for the billing account.
func (b *billingAPI) GetSettlementNoteAdviceByBillingAccount(billingAccountOUID string) (
	[]*SettlementNoteAdvice,
	error,
) {
	url := fmt.Sprintf(b.address+billingPathGetListOfSettlementNodeAdviceByAccount, billingAccountOUID)

	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}
		return nil, NewTriPicaError(
			fmt.Errorf("couldn't retrieve settlement notes with billingAccountOUID %s: %w", billingAccountOUID, err),
		)
	}

	var advices []*SettlementNoteAdvice
	if err := json.Unmarshal(resp.Body(), &advices); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return advices, nil
}

// GetCustomerBillingAccounts returns the list of relevant MBAs and CBAs.
func (b *billingAPI) GetCustomerBillingAccounts(customerOUID string) ([]*BillingAccount, error) {
	url := fmt.Sprintf(b.address+billingPathGetBillingAccountsByCustomer, customerOUID)

	resp, err := b.httpClient.Get(url)
	if err != nil {
		return nil, NewTriPicaError(errors.NewHTTPRequestError(err))
	}

	if resp.StatusCode() != gohttp.StatusOK {
		err := &errors.HTTPError{
			Body:       string(resp.Body()),
			StatusCode: resp.StatusCode(),
		}
		return nil, NewTriPicaError(
			fmt.Errorf("couldn't retrieve customer billing accounts with customerOUID %s: %w", customerOUID, err),
		)
	}

	var billingAccounts []*BillingAccount
	if err := json.Unmarshal(resp.Body(), &billingAccounts); err != nil {
		return nil, NewTriPicaError(errors.NewParseError(err, resp.Body()))
	}

	return billingAccounts, nil
}

const (
	billPresentationMediaPostmail    = "POSTMAIL"
)

// Models for exchanging data with tripica
type (
	BillingAccount struct {
		OUID                        string                       `json:"ouid"`
		Name                        string                       `json:"name"`
		CustomerOUID                string                       `json:"customerOuid"`
		BillPresentationMedia       string                       `json:"billPresentationMedia"`
		DateTimeCreate              Date                         `json:"dateTimeCreate"`
		BillingAccountRelationships []BillingAccountRelationship `json:"billingAccountRelationships"`
	}

	// BillingAccountRelationship represents a triPica billing account relationship.
	BillingAccountRelationship struct {
		OUID                     string `json:"ouid"`
		Type                     string `json:"type"`
		TargetBillingAccountOUID string `json:"targetBillingAccountOuid"`
	}
)

// IsOfflineCustomer checks whether the customer is an offline customer.
func (a *BillingAccount) IsOfflineCustomer() bool {
	return a.BillPresentationMedia == billPresentationMediaPostmail
}

// BalanceType represents a triPica balance type. It is inferred from a list of charges.
type BalanceType string

// Possible balance types.
const (
	BalanceTypeDownPayment           = "Abschlag"
	BalanceTypeBill                  = "Rechnung"
	BalanceTypeDownPaymentAndBankFee = "Abschlag und Bankgebühren"
	BalanceTypeBankFee               = "Bankgebühren"
	BalanceTypeOther                 = "Sonstiges"
)

// BillingAccountBalance represents a triPica billing account balance.
type BillingAccountBalance struct {
	OUID                     string                  `json:"ouid"`
	BillingAccountOUID       string                  `json:"billingAccountOuid"`
	Amount                   int                     `json:"amount"`
	Status                   string                  `json:"status"`
	Type                     string                  `json:"type"`
	TransactionID            string                  `json:"transactionId"`
	SettlementNoteAdviceOUID string                  `json:"settlementNoteAdviceOuid"`
	StartDate                Date                    `json:"startDateTime"`
}

// AppliedBillingCharge represents a triPica applied billing charge.
type AppliedBillingCharge struct {
	OUID               string `json:"ouid"`
	GeneralLedgerID    string `json:"glid"`
	BillingAccountOUID string `json:"billingAccountOuid"`
	TransactionID      string `json:"transactionId"`
	CurrencyCode       string `json:"currencyCode"`
}

// ignoreChargeType checks whether this charge should be ignored when inferring a balance type.
func (c *AppliedBillingCharge) ignoreChargeType() bool {
	chargesToIgnore := [...]string{"CANCELLED", "REJECTED", "RETURN", "REBOOKED", "RECEIVABLE"}
	for _, cti := range chargesToIgnore {
		if strings.Contains(c.GeneralLedgerID, cti) {
			return true
		}
	}
	return false
}

func (c *AppliedBillingCharge) isBankFee() bool {
	return strings.Contains(c.GeneralLedgerID, "BANK_FEE")
}

func (c *AppliedBillingCharge) isDownPayment() bool {
	return strings.Contains(c.GeneralLedgerID, "ABSCHLAG")
}

func (c *AppliedBillingCharge) isBill() bool {
	return strings.Contains(c.GeneralLedgerID, "BILL")
}

// SettlementNoteAdvice represents a triPica settlement note advice.
type SettlementNoteAdvice struct {
	OUID           string `json:"ouid"`
	ID             string `json:"id"`
	BillDate       Date   `json:"billDate"`
	PaymentDueDate int64  `json:"paymentDueDate"`
	Category       string `json:"category"`
	State          string `json:"state"`
}
