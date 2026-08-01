package http

// RetCode is a platform return code.
type RetCode int

const (
	RetOK RetCode = 0

	// The platform authentication block, 1000..1034.
	RetAuthClientInvalid   RetCode = 1000 // invalid type of the terminal
	RetAuthAccountInvalid  RetCode = 1001 // invalid account, our wrong password case
	RetAuthAccountDisabled RetCode = 1002 // the account is disabled
	RetAuthManagerNoConfig RetCode = 1011 // no manager configuration for the account
	RetAuthManagerIpBlock  RetCode = 1012 // ip address is not valid for the manager
	RetAuthServerBusy      RetCode = 1018 // the server is busy
	RetAuthAccountUnknown  RetCode = 1020 // unknown account
	RetAuthManagerType     RetCode = 1024 // this connection type is not permitted for manager
	RetAuthResetPassword   RetCode = 1026 // master password must be changed
	RetAuthOtpInvalid      RetCode = 1027 // invalid one time password
	RetAuthApiDisabled     RetCode = 1034 // api connection prohibited, USER_RIGHT_API_ENABLED

	// hst extensions.
	RetSessionExpired RetCode = 60001 // the session is gone, refresh or log in again
	RetAccountLocked  RetCode = 60002 // too many failed attempts, temporarily locked
)
