package http

import "github.com/gofiber/fiber/v2"

// RetCode is an MT5 EnMTAPIRetcode value. It travels in the response body so a
// client can tell "wrong password" from "must change password" without parsing
// text, while the HTTP status stays in the header.
type RetCode int

const (
	RetOK RetCode = 0

	// authentication block, MT5 1000..1027
	RetAuthClientInvalid   RetCode = 1000
	RetAuthAccountInvalid  RetCode = 1001
	RetAuthAccountDisabled RetCode = 1002
	RetAuthAccountBlocked  RetCode = 1003
	RetAuthServerBusy      RetCode = 1004
	RetAuthManagerInvalid  RetCode = 1011
	RetAuthManagerIpBlock  RetCode = 1012
	RetAuthTerminalInvalid RetCode = 1024
	RetAuthUpdatePassword  RetCode = 1026
	RetAuthSessionExpired  RetCode = 1027
)

// RetHTTPStatus maps a retcode onto the status a client should see.
// A must-change-password login still succeeded, so it stays 200.
func RetHTTPStatus(r RetCode) int {
	switch r {
	case RetOK, RetAuthUpdatePassword:
		return fiber.StatusOK
	case RetAuthClientInvalid, RetAuthAccountInvalid, RetAuthSessionExpired:
		return fiber.StatusUnauthorized
	case RetAuthAccountDisabled, RetAuthAccountBlocked, RetAuthManagerInvalid,
		RetAuthManagerIpBlock, RetAuthTerminalInvalid:
		return fiber.StatusForbidden
	case RetAuthServerBusy:
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
