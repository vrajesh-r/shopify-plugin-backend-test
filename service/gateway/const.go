package gateway

const (
	GATEWAY_PORTAL_COOKIE_NAME   = "bread-gateway-portal-cookie"
	HASH_COST                    = 10
	defaultErrMessage            = "There was an issue authorizing your transaction."
	defaultErrMessageAction      = "Please contact customer support or choose another payment method."
	remainderPayErrMessage       = "The credit/debit card portion of your transaction was declined."
	remainderPayErrMessageAction = "Please use a different card or contact your bank. Otherwise, you can still check out with an amount covered by your Bread loan capacity."
	BreadClassic                 = "classic"
	BreadPlatform                = "platform"
)

var invalidStateCodes = map[string]struct{}{
	"aa": {},
	"ae": {},
	"ap": {},
	"pw": {},
	"as": {},
	"fm": {},
	"gu": {},
	"mh": {},
	"mp": {},
	"pr": {},
}
