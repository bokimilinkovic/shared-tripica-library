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

// GetAssociatedBillingAccounts retrieves all the billing accounts associated to provided master billing account,
// including the MBA itself.
func (b *billingAPI) GetAssociatedBillingAccounts(mba *BillingAccount) ([]*BillingAccount, error) {
	billingAccounts, err := b.GetCustomerBillingAccounts(mba.CustomerOUID)
	if err != nil {
		return nil, err
	}

	return RelevantBillingAccounts(billingAccounts, mba), nil
}

// CustomerHasFinalBill checks whether a final bill is associated with provided billing account.
func (b *billingAPI) CustomerHasFinalBill(billingAccountOUID string) (bool, error) {
	advices, err := b.GetSettlementNoteAdviceByBillingAccount(billingAccountOUID)
	if err != nil {
		return false, err
	}

	return AdvicesContainFinalBill(advices), nil
}

// CustomerWithinGracePeriod checks whether a billing account tied to a customer is still in grace period.
func (b *billingAPI) CustomerWithinGracePeriod(billingAccount *BillingAccount, now time.Time) bool {
	gracePeriod := billingAccount.DateTimeCreate.AddDate(0, 0, int(b.customerGracePeriodDays))
	return now.Before(gracePeriod)
}

// GetOverdueBalances consolidates billing charges and returns a list of overdue balances.
// nolint: gocyclo
func (b *billingAPI) GetOverdueBalances(
	customerOUID string,
	masterBillingAccount *BillingAccount,
) (
	[]*BillingAccountBalance,
	error,
) {
	billingAccounts, err := b.GetCustomerBillingAccounts(customerOUID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get related billing accounts: %w", err)
	}
	billingAccounts = RelevantBillingAccounts(billingAccounts, masterBillingAccount)

	balances, err := b.GetDueBillingAccountBalancesByCustomer(customerOUID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get due account balances: %w", err)
	}
	balances = RelevantBalances(balances, billingAccounts)
	balances = DeduplicateBalancesForTransactions(balances)

	if len(balances) == 0 {
		return balances, nil
	}

	transactionIDs := BillingAccountBalanceTransactionIDs(balances)
	charges, err := b.GetAppliedBillingChargesByTransactionIDs(transactionIDs)
	if err != nil {
		return nil, fmt.Errorf("couldn't get billing charges: %w", err)
	}

	overdueBalances := balances[:0]
	for _, balance := range balances {
		for _, charge := range charges {
			if charge.TransactionID == balance.TransactionID {
				balance.Charges = append(balance.Charges, charge)
			}
		}

		// [ED4FTR-3710] We shouldn't include balances without charges, since they will result in "Sonstiges".
		if len(balance.Charges) == 0 {
			b.logger.Warnf("service/tripica: couldn't find charges for balance with OUID: %s; transaction ID: %s",
				balance.OUID,
				balance.TransactionID,
			)
			continue
		}

		if err := b.inferBalanceData(balance, masterBillingAccount); err != nil {
			return nil, err
		}

		if !balance.Ignore {
			overdueBalances = append(overdueBalances, balance)
		}
	}

	return overdueBalances, nil
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

// AdvicesContainFinalBill checks whether advices contain a final bill.
func AdvicesContainFinalBill(advices []*SettlementNoteAdvice) bool {
	for _, advice := range advices {
		if advice.Category == settlementNoteAdviceCategoryLast && advice.State == settlementNoteAdviceStateSettled {
			return true
		}
	}

	return false
}

// RelevantBillingAccounts goes through the billing accounts, and searches for relevant CBAs.
// CBA is every billing account that has a billing account relationship which targets the master billing account.
func RelevantBillingAccounts(
	billingAccounts []*BillingAccount,
	masterBillingAccount *BillingAccount,
) []*BillingAccount {
	relevantBillingAccounts := billingAccounts[:0]
	for _, a := range billingAccounts {
		for _, r := range a.BillingAccountRelationships {
			if r.Type == billingAccountRelationshipParent && r.TargetBillingAccountOUID == masterBillingAccount.OUID {
				relevantBillingAccounts = append(relevantBillingAccounts, a)
				break
			}
		}
	}
	relevantBillingAccounts = append(relevantBillingAccounts, masterBillingAccount)
	return relevantBillingAccounts
}

// RelevantBalances considers only the balances which are tied to relevant MBA and CBAs. This function assumes
// that the provided array of billing accounts contains only the relevant MBAs and CBAs.
func RelevantBalances(balances []*BillingAccountBalance, billingAccounts []*BillingAccount) []*BillingAccountBalance {
	relevantBalances := balances[:0]
	for _, b := range balances {
		for _, a := range billingAccounts {
			if b.BillingAccountOUID == a.OUID {
				relevantBalances = append(relevantBalances, b)
				break
			}
		}
	}
	return relevantBalances
}

// DeduplicateBalancesForTransactions merges balances with same transaction ID by summing their amounts,
// and then leaving only balances with unique transaction IDs in the list.
func DeduplicateBalancesForTransactions(balances []*BillingAccountBalance) []*BillingAccountBalance {
	const nextIndexIncrement = 1

	merged := map[*BillingAccountBalance]bool{}
	for i := 0; i < len(balances)-1; i++ {
		for j := i + nextIndexIncrement; j < len(balances); j++ {
			if merged[balances[j]] || merged[balances[i]] || balances[i].TransactionID != balances[j].TransactionID {
				continue
			}

			balances[i].Amount += balances[j].Amount
			merged[balances[j]] = true
		}
	}

	mergedBalances := balances[:0]
	for _, balance := range balances {
		if !merged[balance] {
			mergedBalances = append(mergedBalances, balance)
		}
	}

	return mergedBalances
}

// BillingAccountBalanceTransactionIDs returns a comma separated list of transaction IDs.
func BillingAccountBalanceTransactionIDs(balances []*BillingAccountBalance) string {
	transactionIDs := []string{}
	for _, b := range balances {
		transactionIDs = append(transactionIDs, b.TransactionID)
	}
	return strings.Join(transactionIDs, ",")
}
