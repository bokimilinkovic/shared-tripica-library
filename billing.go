package tripica

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"shared-tripica-library/http"
	"shared-tripica-library/http/errors"
	"shared-tripica-library/log"
	"strings"
	"time"
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

	// Defines the amount of days in the future, after which
	// a non-bill transaction is due.
	dueDateDays uint

	// Defines the amount of days since the creation of the customer's billing account
	// before we start creating any claims for them.
	customerGracePeriodDays uint

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


func (b *billingAPI) inferBalanceData(balance *BillingAccountBalance, billingAccount *BillingAccount) error {
	balance.InferBalanceType()

	if balance.InferredBalanceType != BalanceTypeBill {
		balance.InferDueDate(b.dueDateDays)
		return nil
	}

	// If balance type is BILL, then we need to fetch settlement node advices and compare
	// against each to infer the due date.
	advices, err := b.GetSettlementNoteAdviceByBillingAccount(billingAccount.OUID)
	if err != nil {
		return fmt.Errorf("couldn't get settlement note advices: %w", err)
	}

	dueDateInferred := false
	for _, advice := range advices {
		if balance.SettlementNoteAdviceOUID == advice.OUID {
			balance.SettlementNoteAdvice = advice
			balance.InferBillDueDate(advice)
			dueDateInferred = true
			break
		}
	}

	if !dueDateInferred {
		// If no settlement advice could be associated with this balance, then ignore it.
		balance.Ignore = true
		b.logger.Warnf("service/tripica: couldn't infer due date for a bill; balance OUID: %s; transaction ID: %s",
			balance.OUID,
			balance.TransactionID,
		)
	}

	return nil
}

const (
	billPresentationMediaPostmail    = "POSTMAIL"
	settlementNoteAdviceCategoryLast = "LAST"
	settlementNoteAdviceStateSettled = "SETTLED"
	billingAccountRelationshipParent = "PARENT"
)

//Models for exchanging data with tripica
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
	SettlementNoteAdvice     *SettlementNoteAdvice   `json:"-"` // Calculated by Dunning Coordinator.
	DueDate                  Date                    `json:"-"` // Calculated by Dunning Coordinator.
	Ignore                   bool                    `json:"-"` // Calculated by Dunning Coordinator.
	InferredBalanceType      BalanceType             `json:"-"` // Calculated by Dunning Coordinator.
	Charges                  []*AppliedBillingCharge `json:"-"` // Calculated by Dunning Coordinator.
}

// InferBalanceType infers balance types from related charges.
// There are 5 possible types a balance can have:
// 	1. Abschlag - if only abschlag charges are present
// 	2. Rechnung - if any bill charges are present
// 	3. Abschlag und Bankgebühren - if both down payments and bank fees are present
// 	4. Bankgebühren - if only bank fee is present
// 	5. Other - if no balance type can be inferred from charges.
func (b *BillingAccountBalance) InferBalanceType() {
	isDownPayment, isBill, isBankFee := false, false, false

	for _, c := range b.Charges {
		if c.ignoreChargeType() {
			continue
		}
		if c.isBill() {
			isBill = true
			// If at least one bill is contained, then we don't need to filter the others - the balance type is a bill.
			break
		}
		if c.isDownPayment() {
			isDownPayment = true
		}

		if c.isBankFee() {
			isBankFee = true
		}
	}

	b.inferTypeFromContainedCharges(isDownPayment, isBill, isBankFee)
}

// InferBillDueDate infers the due date for a bill balance, based on the charges.
func (b *BillingAccountBalance) InferBillDueDate(settlementNoteAdvice *SettlementNoteAdvice) {
	if b.InferredBalanceType != BalanceTypeBill {
		return
	}

	today := time.Now().UTC()
	paymentDueDate := time.Unix(0, settlementNoteAdvice.PaymentDueDate*int64(time.Millisecond))
	if today.Before(paymentDueDate) {
		// Ignore balances with dates in the future.
		b.Ignore = true
		return
	}

	b.DueDate.Time = paymentDueDate
}

// InferDueDate infers the due date for a non-bill balance.
func (b *BillingAccountBalance) InferDueDate(dueDateDays uint) {
	if b.InferredBalanceType == BalanceTypeBill {
		return
	}

	today := time.Now().UTC()
	calculatedDueDate := b.StartDate.Time.AddDate(0, 0, int(dueDateDays))
	if today.Before(calculatedDueDate) {
		// Ignore balances with dates in the future.
		b.Ignore = true
		return
	}

	b.DueDate.Time = b.StartDate.Time
}

func (b *BillingAccountBalance) inferTypeFromContainedCharges(downPayment, bill, bankFee bool) {
	switch {
	case bill:
		b.InferredBalanceType = BalanceTypeBill
	case bankFee && downPayment:
		b.InferredBalanceType = BalanceTypeDownPaymentAndBankFee
	case downPayment:
		b.InferredBalanceType = BalanceTypeDownPayment
	case bankFee:
		b.InferredBalanceType = BalanceTypeBankFee
	default:
		b.InferredBalanceType = BalanceTypeOther
	}
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
