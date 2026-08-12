package journal

import "fmt"

// Journal messages. The acting login lives in the login column, so the text only names the record.
var (

	// What a trader does with their own account, written in the words the terminal uses.
	SignedInMsg  = func(ip string) string { return fmt.Sprintf("signed in from %s", ip) }
	SignedOutMsg = func() string { return "signed out" }

	OrderAskedMsg = func(kind, symbol string, lots float64) string {
		return fmt.Sprintf("%s order for %.2f lots of %s was requested", kind, lots, symbol)
	}
	OrderChangedMsg   = func(order int64) string { return fmt.Sprintf("order #%d was modified", order) }
	OrderCancelledMsg = func(order int64) string { return fmt.Sprintf("order #%d was cancelled", order) }

	PositionStopsMsg = func(position int64, sl, tp float64) string {
		return fmt.Sprintf("stop levels on position #%d were set to sl %g and tp %g", position, sl, tp)
	}
	PositionClosedMsg = func(position int64, lots float64) string {
		if lots <= 0 {
			return fmt.Sprintf("close of position #%d was requested", position)
		}
		return fmt.Sprintf("close of %.2f lots on position #%d was requested", lots, position)
	}
	PositionClosedByMsg = func(position, by int64) string {
		return fmt.Sprintf("position #%d was closed by position #%d", position, by)
	}

	MailSentMsg  = func(subject string) string { return fmt.Sprintf("mail %q was sent", subject) }
	MailDraftMsg = func(subject string) string { return fmt.Sprintf("mail draft %q was saved", subject) }

	PasswordChangedMsg = func() string { return "account password was changed" }

	RegisteredMsg = func(accountType string) string {
		return fmt.Sprintf("new %s account was registered", accountType)
	}

	GroupCreatedMsg = func(path string) string { return fmt.Sprintf("group '%s' was created", path) }
	GroupUpdatedMsg = func(path string) string { return fmt.Sprintf("group '%s' configuration was updated", path) }
	GroupDeletedMsg = func(path string) string { return fmt.Sprintf("group '%s' was deleted", path) }

	EndOfDayTimeMsg = func(at string) string {
		return fmt.Sprintf("end of day processing was rescheduled to %s", at)
	}
	EndOfDayRunMsg = func() string { return "end of day processing was started manually" }

	GroupSymbolCreatedMsg = func(path string) string { return fmt.Sprintf("group symbol '%s' was created", path) }
	GroupSymbolUpdatedMsg = func(path string) string { return fmt.Sprintf("group symbol '%s' was updated", path) }
	GroupSymbolDeletedMsg = func(path string) string { return fmt.Sprintf("group symbol '%s' was deleted", path) }

	UserCreatedMsg = func(user int64) string { return fmt.Sprintf("trading account #%d was created", user) }
	UserUpdatedMsg = func(user int64) string { return fmt.Sprintf("trading account #%d was updated", user) }
	UserDeletedMsg = func(user int64) string { return fmt.Sprintf("trading account #%d was deleted", user) }
	UserMovedMsg   = func(user int64) string { return fmt.Sprintf("trading account #%d was moved to another group", user) }

	ClientCreatedMsg = func(client int64) string { return fmt.Sprintf("client #%d was created", client) }
	ClientUpdatedMsg = func(client int64) string { return fmt.Sprintf("client #%d was updated", client) }
	ClientDeletedMsg = func(client int64) string { return fmt.Sprintf("client #%d was deleted", client) }

	ManagerCreatedMsg = func(manager int64) string { return fmt.Sprintf("manager account #%d was created", manager) }
	ManagerUpdatedMsg = func(manager int64) string { return fmt.Sprintf("manager account #%d was updated", manager) }
	ManagerDeletedMsg = func(manager int64) string { return fmt.Sprintf("manager account #%d was deleted", manager) }
	ManagerGrantedMsg = func(manager int64) string {
		return fmt.Sprintf("group access of manager account #%d was granted", manager)
	}

	GroupCommissionCreatedMsg = func(path string) string {
		return fmt.Sprintf("commission for group '%s' was created", path)
	}
	GroupCommissionUpdatedMsg = func(path string) string {
		return fmt.Sprintf("commission for group '%s' was updated", path)
	}
	GroupCommissionDeletedMsg = func(path string) string {
		return fmt.Sprintf("commission for group '%s' was deleted", path)
	}

	SymbolCreatedMsg = func(symbol string) string { return fmt.Sprintf("symbol '%s' was created", symbol) }
	SymbolUpdatedMsg = func(symbol string) string { return fmt.Sprintf("symbol '%s' was updated", symbol) }
	SymbolDeletedMsg = func(symbol string) string { return fmt.Sprintf("symbol '%s' was deleted", symbol) }

	MailServerCreatedMsg = func(id int) string { return fmt.Sprintf("mail server #%d was created", id) }
	MailServerUpdatedMsg = func(id int) string { return fmt.Sprintf("mail server #%d was updated", id) }
	MailServerDeletedMsg = func(id int) string { return fmt.Sprintf("mail server #%d was deleted", id) }

	HolidayCreatedMsg   = func(id int) string { return fmt.Sprintf("holiday #%d was created", id) }
	HolidayUpdatedMsg   = func(id int) string { return fmt.Sprintf("holiday #%d was updated", id) }
	HolidayDeletedMsg   = func(id int) string { return fmt.Sprintf("holiday #%d was deleted", id) }
	HolidayReorderedMsg = func() string { return "the holiday list was reordered" }

	LeverageCreatedMsg = func(name string) string {
		return fmt.Sprintf("leverage profile '%s' was created", name)
	}
	LeverageUpdatedMsg = func(name string) string {
		return fmt.Sprintf("leverage profile '%s' was updated", name)
	}
	LeverageDeletedMsg     = func(id int) string { return fmt.Sprintf("leverage profile #%d was deleted", id) }
	LeverageRuleCreatedMsg = func(id int) string { return fmt.Sprintf("leverage rule #%d was created", id) }
	LeverageRuleUpdatedMsg = func(id int) string { return fmt.Sprintf("leverage rule #%d was updated", id) }
	LeverageRuleDeletedMsg = func(id int) string { return fmt.Sprintf("leverage rule #%d was deleted", id) }
	LeverageReorderedMsg   = func(id int) string {
		return fmt.Sprintf("rules of leverage profile #%d were reordered", id)
	}

	DatafeedConnectedMsg = func(name string) string {
		return fmt.Sprintf("datafeed '%s' established a connection to its source", name)
	}
	DatafeedDisconnectedMsg = func(name string) string {
		return fmt.Sprintf("datafeed '%s' lost the connection to its source", name)
	}
)

var BalanceMsg = func(login int64, action string, amount float64) string {
	return fmt.Sprintf("%s of %.2f was applied to account #%d", action, amount, login)
}

// SessionDisconnectedMsg says a manager cut one of the account's live sessions.
func SessionDisconnectedMsg(login int64) string {
	return fmt.Sprintf("a session of account #%d was disconnected by the desk", login)
}

func PositionInvalidMsg(positionId int64, volume, validVolume, price, validPrice float64) string {
	return fmt.Sprintf("position #%d has invalid volume: %.2f, valid: %.2f, price: %.5f, valid: %.5f",
		positionId, volume, validVolume, price, validPrice)
}

func PositionFixedMsg(positionId int64, from, to float64) string {
	return fmt.Sprintf("position #%d was fixed from volume %.2f to %.2f", positionId, from, to)
}

func PositionDeletedMsg(positionId, login int64) string {
	return fmt.Sprintf("position #%d of account #%d was deleted", positionId, login)
}

func DealUpdatedMsg(dealId, login int64, volFrom, volTo, profitFrom, profitTo float64) string {
	return fmt.Sprintf("deal #%d of account #%d was updated: volume %.2f -> %.2f, profit %.2f -> %.2f",
		dealId, login, volFrom, volTo, profitFrom, profitTo)
}

func DealDeletedMsg(dealId, login int64) string {
	return fmt.Sprintf("deal #%d of account #%d was deleted", dealId, login)
}
