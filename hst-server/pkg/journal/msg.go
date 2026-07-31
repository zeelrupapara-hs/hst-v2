package journal

import "fmt"

// Journal messages. The first argument is the manager who acted, the second names the record; the rest is in the detail.
var (
	GroupCreatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group created", login, path) }
	GroupUpdatedMsg = func(login int64, path string) string { return fmt.Sprintf("%d: %s group updated", login, path) }
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
)
