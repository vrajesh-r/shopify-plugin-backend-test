package bread

import (
	"fmt"
)

func TransactionActionUrl(txId, ostiaHost string) string {
	return ostiaHost + "/transactions/actions/" + txId
}

func QueryTransactionUrl(txId, ostiaHost string) string {
	return ostiaHost + "/transactions/" + txId
}

func SaveCartUrl(ostiaHost string) string {
	return ostiaHost + "/carts/"
}

func SendCartEmailUrl(ostiaHost, cartID string) string {
	return ostiaHost + "/carts/" + cartID + "/email"
}

func SendCartTextUrl(ostiaHost, cartID string) string {
	return ostiaHost + "/carts/" + cartID + "/text"
}

func ExpireCartUrl(ostiaHost, cartID string) string {
	return ostiaHost + "/carts/" + cartID + "/expire"
}

func TransactionShipmentURL(transactionID, ostiaHost string) string {
	// https://api.getbread.com/transactions/:tx-id/shipment
	return ostiaHost + "/transactions/" + transactionID + "/shipment"
}

func AuthorizeTransactionURL(host, trxID string) string {
	return fmt.Sprintf("%s/api/transaction/%s/authorize", host, trxID)
}

func CancelTransactionURL(host, trxID string) string {
	return fmt.Sprintf("%s/api/transaction/%s/cancel", host, trxID)
}

func RefundTransactionURL(host, trxID string) string {
	return fmt.Sprintf("%s/api/transaction/%s/refund", host, trxID)
}

func SettleTransactionURL(host, trxID string) string {
	return fmt.Sprintf("%s/api/transaction/%s/settle", host, trxID)
}

func TransactionAuthTokenURL(host string) string {
	return fmt.Sprintf("%s/api/auth/sa/authenticate", host)
}

func GetTransactionFromApplicationURL(host, applicationID string) string {
	return fmt.Sprintf("%s/api/transaction/application/%s", host, applicationID)
}

func GetTransactionURL(host, transactionID string) string {
	return fmt.Sprintf("%s/api/transaction/%s", host, transactionID)
}
