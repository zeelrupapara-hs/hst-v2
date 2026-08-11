package v1

import (
	"context"
	"strings"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// The back office navigator.
//
// Two panels read this: the administrator sets the platform up, the manager runs it day to day.
// They are different trees, not one tree with things hidden, which is how the platform itself
// draws them. What a caller may see inside either is settled by their rights and by the group
// masks bounding what they may reach, so the server builds the tree rather than the panel.

// Terminal names the panel a session belongs to.
const (
	TerminalAdmin   = "administrator"
	TerminalManager = "manager"
)

// ViewNavNode is one entry of the navigator.
type ViewNavNode struct {
	Key      string        `json:"key"`
	Label    string        `json:"label"`
	Icon     string        `json:"icon"`
	Route    string        `json:"route"`
	Section  string        `json:"section"`
	Count    *int64        `json:"count,omitempty"`
	Children []ViewNavNode `json:"children,omitempty"`
}

// ViewNavigation is the navigator plus what the caller may do once inside it.
type ViewNavigation struct {
	Login    int64  `json:"login"`
	Terminal string `json:"terminal"`
	// Groups are the masks bounding every count and every list this caller sees.
	Groups []string            `json:"groups"`
	Nodes  []ViewNavNode       `json:"nodes"`
	Can    map[string]bool     `json:"can"`
	Rights model.ManagerRights `json:"rights"`
	// Visible and Total say how much of this terminal's navigator the caller was given.
	Visible int `json:"visible"`
	Total   int `json:"total"`
	Hidden  int `json:"hidden"`
}

// navNode is a navigator entry before the caller's rights are applied.
type navNode struct {
	key, label, icon, route, section string
	// needs is every right the entry requires; the platform states these dependencies and a
	// missing one hides the entry, so orders need account access as well as trade access
	needs    []uint
	counter  string
	children []navNode
}

// Every entry is a section this platform serves, gated on the same rights the route behind it
// enforces, so the navigator can never offer what the api would refuse.

// managerTree is the running of the platform: the accounts, their trades, and the desk.
var managerTree = []navNode{
	{key: "clients_orders", label: "Clients & Orders", icon: "users", route: "", section: "trading",
		children: []navNode{
			{key: "accounts", label: "Trading Accounts", icon: "user", route: "/users", section: "trading",
				needs: []uint{model.MgrRightAccRead}, counter: "accounts"},
			{key: "clients", label: "Clients", icon: "briefcase", route: "/clients", section: "trading",
				needs: []uint{model.MgrRightClientsAccess}, counter: "clients"},
			{key: "positions", label: "Positions", icon: "layers", route: "/positions", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "positions"},
			{key: "orders", label: "Orders", icon: "list", route: "/orders", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "orders"},
			{key: "deals", label: "Deals", icon: "receipt", route: "/deals", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}},
		}},

	{key: "dealing", label: "Dealing", icon: "gavel", route: "/dealing", section: "dealing",
		needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead, model.MgrRightTradesDealer}},

	{key: "margin_calls", label: "Margin Calls", icon: "security", route: "/margin-calls",
		section: "dealing", needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}},

	{key: "market_watch", label: "Market Watch", icon: "symbols", route: "/market",
		section: "trading"},

	{key: "balance", label: "Balance Operations", icon: "wallet", route: "/balance", section: "accounting",
		needs: []uint{model.MgrRightAccountant}},

	{key: "groups", label: "Groups", icon: "folder-tree", route: "/groups", section: "config",
		needs: []uint{model.MgrRightCfgGroups}, counter: "groups"},

	{key: "journal", label: "Journal", icon: "scroll", route: "/journal", section: "support",
		needs: []uint{model.MgrRightSrvJournals}},
}

// adminTree is the setting up of the platform: instruments, groups, feeds and who may run it.
var adminTree = []navNode{
	{key: "clients_accounts", label: "Clients & Accounts", icon: "users", route: "", section: "accounts",
		children: []navNode{
			{key: "accounts", label: "Trading Accounts", icon: "user", route: "/users", section: "accounts",
				needs: []uint{model.MgrRightAccRead}, counter: "accounts"},
			{key: "clients", label: "Clients", icon: "briefcase", route: "/clients", section: "accounts",
				needs: []uint{model.MgrRightClientsAccess}, counter: "clients"},
			{key: "managers", label: "Managers", icon: "shield", route: "/managers", section: "accounts",
				needs: []uint{model.MgrRightCfgManagers}, counter: "managers"},
		}},

	{key: "orders_deals", label: "Orders & Deals", icon: "list", route: "", section: "trading",
		needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead},
		children: []navNode{
			{key: "positions", label: "Positions", icon: "layers", route: "/positions", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "positions"},
			{key: "orders", label: "Orders", icon: "list", route: "/orders", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "orders"},
			{key: "deals", label: "Deals", icon: "receipt", route: "/deals", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}},
		}},

	{key: "holidays", label: "Holidays", icon: "calendar", route: "/holidays", section: "config",
		needs: []uint{model.MgrRightCfgHolidays}, counter: "holidays"},

	{key: "leverages", label: "Leverages", icon: "percent", route: "/leverage-profiles",
		section: "config", needs: []uint{model.MgrRightCfgGroups}, counter: "leverages"},

	{key: "groups", label: "Groups", icon: "folder-tree", route: "/groups", section: "config",
		needs: []uint{model.MgrRightCfgGroups}, counter: "groups"},

	{key: "symbols", label: "Symbols", icon: "tag", route: "/symbols", section: "config",
		needs: []uint{model.MgrRightCfgSymbols}, counter: "symbols"},

	{key: "routing", label: "Routing", icon: "git-branch", route: "/routing", section: "config",
		needs: []uint{model.MgrRightCfgRequests}, counter: "routing"},

	{key: "datafeeds", label: "Data Feeds", icon: "rss", route: "/datafeeds", section: "config",
		needs: []uint{model.MgrRightCfgDatafeeds}, counter: "datafeeds"},

	{key: "mail_servers", label: "Mail Servers", icon: "mail", route: "/mail-servers", section: "feeds",
		needs: []uint{model.MgrRightCfgMails}, counter: "mail_servers"},

	{key: "end_of_day", label: "End of Day", icon: "clock", route: "/system/end-of-day", section: "config",
		needs: []uint{model.MgrRightCfgTime}},

	{key: "journal", label: "Journal", icon: "scroll", route: "/journal", section: "support",
		needs: []uint{model.MgrRightSrvJournals}},
}

// NavigationTree is the navigator for whichever panel the caller signed in to.
//
//	@Id			NavigationTree
//	@Tags		Navigation
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewNavigation}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/navigation [get]
func (s *HttpServer) NavigationTree(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rights := snap.ManagerRights
	terminal := terminalOf(snap.ConnectionType)

	tree := managerTree
	if terminal == TerminalAdmin {
		tree = adminTree
	}

	counts := s.navCounts(c.UserContext(), snap.IsManager, snap.ManagerGroups, rights)
	nodes := permitted(tree, rights, counts)

	out := &ViewNavigation{
		Login:    snap.Login,
		Terminal: terminal,
		Groups:   snap.ManagerGroups,
		Nodes:    nodes,
		Can:      rights.Flags(),
		Rights:   rights,
		Visible:  countNodes(nodes),
		Total:    countDefined(tree),
	}
	out.Hidden = out.Total - out.Visible

	// the groups the masks reach hang under the Groups node, as the platform's navigator does
	if node := findNode(out.Nodes, "groups"); node != nil {
		node.Children = s.groupNodes(c.UserContext(), snap.IsManager, snap.ManagerGroups)
	}

	return s.App.HttpResponseOK(c, out)
}

// permitted keeps the entries the rights allow, dropping a branch that ends up with nothing under it.
func permitted(nodes []navNode, r model.ManagerRights, counts map[string]int64) []ViewNavNode {
	out := []ViewNavNode{}

	for _, n := range nodes {
		if !allows(r, n.needs) {
			continue
		}

		children := permitted(n.children, r, counts)

		// a branch with no route of its own is only a heading, and an empty heading is noise
		if n.route == "" && len(children) == 0 {
			continue
		}

		v := ViewNavNode{Key: n.key, Label: n.label, Icon: n.icon, Route: n.route,
			Section: n.section, Children: children}

		if n.counter != "" {
			if c, ok := counts[n.counter]; ok {
				v.Count = &c
			}
		}

		out = append(out, v)
	}

	return out
}

// allows reports whether every right an entry depends on is held.
func allows(r model.ManagerRights, needs []uint) bool {
	for _, bit := range needs {
		if !r.Has(bit) {
			return false
		}
	}

	return true
}

// countNodes is how many entries the caller was actually given.
func countNodes(nodes []ViewNavNode) int {
	n := len(nodes)
	for i := range nodes {
		n += countNodes(nodes[i].Children)
	}

	return n
}

// countDefined is how many entries this terminal's navigator has in total.
func countDefined(nodes []navNode) int {
	n := len(nodes)
	for i := range nodes {
		n += countDefined(nodes[i].children)
	}

	return n
}

// findNode walks the built tree for one key.
func findNode(nodes []ViewNavNode, key string) *ViewNavNode {
	for i := range nodes {
		if nodes[i].Key == key {
			return &nodes[i]
		}
		if n := findNode(nodes[i].Children, key); n != nil {
			return n
		}
	}

	return nil
}

// navCounts is what each section holds for this caller, counted through their group masks.
func (s *HttpServer) navCounts(ctx context.Context, isManager bool, masks []string,
	r model.ManagerRights) map[string]int64 {
	out := make(map[string]int64, 8)

	where, args := utils.GroupAccessFor(isManager, masks, `u."group"`, 1)

	if r.Has(model.MgrRightAccRead) {
		out["accounts"] = s.countOf(ctx, `SELECT count(*) FROM hst.users u WHERE `+where, args)

		if r.Has(model.MgrRightTradesRead) {
			out["positions"] = s.countOf(ctx,
				`SELECT count(*) FROM hst.positions p JOIN hst.users u ON u.login = p.login WHERE `+where, args)
			out["orders"] = s.countOf(ctx,
				`SELECT count(*) FROM hst.orders o JOIN hst.users u ON u.login = o.login
				  WHERE o.state IN (0, 1, 3, 7, 8, 9) AND `+where, args)
		}
	}

	if r.Has(model.MgrRightClientsAccess) {
		out["clients"] = s.countOf(ctx, `SELECT count(*) FROM hst.clients`, nil)
	}
	if r.Has(model.MgrRightCfgManagers) {
		out["managers"] = s.countOf(ctx, `SELECT count(*) FROM hst.managers`, nil)
	}
	if r.Has(model.MgrRightCfgDatafeeds) {
		out["datafeeds"] = s.countOf(ctx, `SELECT count(*) FROM hst.datafeeds`, nil)
	}
	if r.Has(model.MgrRightCfgMails) {
		out["mail_servers"] = s.countOf(ctx, `SELECT count(*) FROM hst.mail_servers`, nil)
	}

	if r.Has(model.MgrRightCfgHolidays) {
		out["holidays"] = s.countOf(ctx, `SELECT count(*) FROM hst.holidays`, nil)
	}
	if r.Has(model.MgrRightCfgGroups) {
		out["groups"] = s.countOf(ctx, `SELECT count(*) FROM hst.groups`, nil)
		out["leverages"] = s.countOf(ctx, `SELECT count(*) FROM hst.leverages`, nil)
	}
	if r.Has(model.MgrRightCfgSymbols) {
		out["symbols"] = s.countOf(ctx, `SELECT count(*) FROM hst.symbols`, nil)
	}
	if r.Has(model.MgrRightCfgRequests) {
		out["routing"] = s.countOf(ctx, `SELECT count(*) FROM hst.routing`, nil)
	}

	return out
}

// countOf answers zero rather than failing the whole navigator on one bad count.
func (s *HttpServer) countOf(ctx context.Context, query string, args []any) int64 {
	var n int64
	if err := s.DB.DB.QueryRow(ctx, query, args...).Scan(&n); err != nil {
		return 0
	}

	return n
}

// groupNodes is every group the caller's masks reach, with how many accounts sit in each.
func (s *HttpServer) groupNodes(ctx context.Context, isManager bool, masks []string) []ViewNavNode {
	where, args := utils.GroupAccessFor(isManager, masks, `g."group"`, 1)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT g."group", count(u.login)
		   FROM hst.groups g
		   LEFT JOIN hst.users u ON u."group" = g."group"
		  WHERE `+where+`
		  GROUP BY g."group"
		  ORDER BY g."group"`, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := []ViewNavNode{}

	for rows.Next() {
		var name string
		var n int64
		if err := rows.Scan(&name, &n); err != nil {
			return out
		}

		count := n
		out = append(out, ViewNavNode{
			Key:     "group:" + name,
			Label:   leafOf(name),
			Icon:    "folder",
			Route:   "/groups/" + name,
			Section: "config",
			Count:   &count,
		})
	}

	return out
}

// leafOf is the last part of a group path, which is what the navigator shows.
func leafOf(group string) string {
	if i := strings.LastIndex(group, `\`); i >= 0 && i+1 < len(group) {
		return group[i+1:]
	}

	return group
}

// terminalOf names the panel a connection type signed in to.
func terminalOf(t int32) string {
	switch model.UsersConnectionTypes(t) {
	case model.UsersConnectionTypes_admin, model.UsersConnectionTypes_admin_api:
		return TerminalAdmin
	default:
		return TerminalManager
	}
}
