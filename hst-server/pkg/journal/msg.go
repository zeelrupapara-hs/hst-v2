package journal

import "fmt"

// Journal messages. The first argument is the manager who acted, the second names the record; the rest is in the detail.
var (
	RegisteredMsg = func(login int64, accountType string) string {
		return fmt.Sprintf("%d: %s account registered", login, accountType)
	}

	GroupCreatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group created", login, path) }
	GroupUpdatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group updated", login, path) }

	EndOfDayTimeMsg = func(login int64, at string) string {
		return fmt.Sprintf("%d: end of day moved to %s", login, at)
	}
	EndOfDayRunMsg  = func(login int64) string { return fmt.Sprintf("%d: end of day run by hand", login) }
	GroupDeletedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group deleted", login, path) }

	GroupSymbolCreatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group symbol created", login, path) }
	GroupSymbolUpdatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group symbol updated", login, path) }
	GroupSymbolDeletedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group symbol deleted", login, path) }

	UserCreatedMsg = func(login, user int64) string { return fmt.Sprintf("%d: %d user created", login, user) }
	UserUpdatedMsg = func(login, user int64) string { return fmt.Sprintf("%d: %d user updated", login, user) }
	UserDeletedMsg = func(login, user int64) string { return fmt.Sprintf("%d: %d user deleted", login, user) }
	UserMovedMsg   = func(login, user int64) string { return fmt.Sprintf("%d: %d user moved", login, user) }

	ClientCreatedMsg = func(login, client int64) string { return fmt.Sprintf("%d: %d client created", login, client) }
	ClientUpdatedMsg = func(login, client int64) string { return fmt.Sprintf("%d: %d client updated", login, client) }
	ClientDeletedMsg = func(login, client int64) string { return fmt.Sprintf("%d: %d client deleted", login, client) }

	ManagerCreatedMsg = func(login, manager int64) string { return fmt.Sprintf("%d: %d manager created", login, manager) }
	ManagerUpdatedMsg = func(login, manager int64) string { return fmt.Sprintf("%d: %d manager updated", login, manager) }
	ManagerDeletedMsg = func(login, manager int64) string { return fmt.Sprintf("%d: %d manager deleted", login, manager) }
	ManagerGrantedMsg = func(login, manager int64) string { return fmt.Sprintf("%d: %d manager group granted", login, manager) }

	GroupCommissionCreatedMsg = func(login int64, path string) string {
		return fmt.Sprintf("%d: %s group commission created", login, path)
	}
	GroupCommissionUpdatedMsg = func(login int64, path string) string {
		return fmt.Sprintf("%d: %s group commission updated", login, path)
	}
	GroupCommissionDeletedMsg = func(login int64, path string) string {
		return fmt.Sprintf("%d: %s group commission deleted", login, path)
	}

	SymbolCreatedMsg = func(login int64, symbol string) string { return fmt.Sprintf("%d: %s symbol created", login, symbol) }
	SymbolUpdatedMsg = func(login int64, symbol string) string { return fmt.Sprintf("%d: %s symbol updated", login, symbol) }
	SymbolDeletedMsg = func(login int64, symbol string) string { return fmt.Sprintf("%d: %s symbol deleted", login, symbol) }

	HolidayCreatedMsg   = func(login int64, id int) string { return fmt.Sprintf("%d: %d holiday created", login, id) }
	HolidayUpdatedMsg   = func(login int64, id int) string { return fmt.Sprintf("%d: %d holiday updated", login, id) }
	HolidayDeletedMsg   = func(login int64, id int) string { return fmt.Sprintf("%d: %d holiday deleted", login, id) }
	HolidayReorderedMsg = func(login int64) string { return fmt.Sprintf("%d: holidays reordered", login) }

	LeverageCreatedMsg = func(login int64, name string) string {
		return fmt.Sprintf("%d: %s leverage profile created", login, name)
	}
	LeverageUpdatedMsg = func(login int64, name string) string {
		return fmt.Sprintf("%d: %s leverage profile updated", login, name)
	}
	LeverageDeletedMsg     = func(login int64, id int) string { return fmt.Sprintf("%d: %d leverage profile deleted", login, id) }
	LeverageRuleCreatedMsg = func(login int64, id int) string { return fmt.Sprintf("%d: %d leverage rule created", login, id) }
	LeverageRuleUpdatedMsg = func(login int64, id int) string { return fmt.Sprintf("%d: %d leverage rule updated", login, id) }
	LeverageRuleDeletedMsg = func(login int64, id int) string { return fmt.Sprintf("%d: %d leverage rule deleted", login, id) }
	LeverageReorderedMsg   = func(login int64, id int) string { return fmt.Sprintf("%d: %d leverage rules reordered", login, id) }
)

var BalanceMsg = func(actor, login int64, action string, amount float64) string {
	return fmt.Sprintf("%d: %s %.2f on %d", actor, action, amount, login)
}
