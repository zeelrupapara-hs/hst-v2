package journal

import (
	"fmt"
	"strings"
)

// Message templates. <login> is the manager who acted, the rest name the record.
const (
	GroupCreatedMsg = "<login>: <grouppath> group created"
	GroupUpdatedMsg = "<login>: <grouppath> group updated"
	GroupDeletedMsg = "<login>: <grouppath> group deleted"

	GroupSymbolCreatedMsg = "<login>: <symbol> group symbol created in <grouppath>"
	GroupSymbolUpdatedMsg = "<login>: <symbol> group symbol updated in <grouppath>"
	GroupSymbolDeletedMsg = "<login>: <symbol> group symbol deleted from <grouppath>"

	UserCreatedMsg = "<login>: <user> user created in <grouppath>"
	UserUpdatedMsg = "<login>: <user> user updated"
	UserDeletedMsg = "<login>: <user> user deleted from <grouppath>"
	UserMovedMsg   = "<login>: <user> user moved from <grouppath> to <topath>"

	ClientCreatedMsg = "<login>: <client> client created"
	ClientUpdatedMsg = "<login>: <client> client updated"
	ClientDeletedMsg = "<login>: <client> client deleted"

	ManagerCreatedMsg      = "<login>: <manager> manager created"
	ManagerUpdatedMsg      = "<login>: <manager> manager updated"
	ManagerDeletedMsg      = "<login>: <manager> manager deleted"
	ManagerGroupGrantedMsg = "<login>: <grouppath> group access granted to <manager> manager"
)

// Msg fills the <name> placeholders of a template from name and value pairs.
func Msg(template string, pairs ...any) string {
	if len(pairs)%2 != 0 {
		return template
	}

	replace := make([]string, 0, len(pairs))
	for i := 0; i < len(pairs); i += 2 {
		replace = append(replace, "<"+fmt.Sprint(pairs[i])+">", fmt.Sprint(pairs[i+1]))
	}

	return strings.NewReplacer(replace...).Replace(template)
}
