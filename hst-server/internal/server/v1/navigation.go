package v1

import (
	"context"
	"strings"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewNavNode is one entry of the back office navigator.
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
	Login    int64               `json:"login"`
	Terminal string              `json:"terminal"`
	Groups   []string            `json:"groups"`
	Nodes    []ViewNavNode       `json:"nodes"`
	Can      map[string]bool     `json:"can"`
	Rights   model.ManagerRights `json:"rights"`
}

// navNode is a navigator entry before the caller's rights are applied.
type navNode struct {
	key, label, icon, route, section string
	// needs is every right the entry requires; MT5 states these dependencies, a missing one hides it
	needs    []uint
	counter  string
	children []navNode
}

// navTree mirrors the MetaTrader 5 manager Navigator and Window menu.
var navTree = []navNode{
	{key: "reports", label: "Reports", icon: "file-chart", route: "/reports", section: "reports",
		needs: []uint{model.MgrRightReports}},

	{key: "clients_orders", label: "Clients & Orders", icon: "users", route: "", section: "trading",
		needs: []uint{model.MgrRightAccRead},
		children: []navNode{
			{key: "online_users", label: "Online Users", icon: "user-check", route: "/online", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightAccOnline}, counter: "online"},
			{key: "accounts", label: "Trading Accounts", icon: "user", route: "/accounts", section: "trading",
				needs: []uint{model.MgrRightAccRead}, counter: "accounts"},
			// orders and positions need account access too, the platform checks both
			{key: "positions", label: "Positions", icon: "layers", route: "/positions", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "positions"},
			{key: "orders", label: "Orders", icon: "list", route: "/orders", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}, counter: "orders"},
			{key: "deals", label: "Deals", icon: "receipt", route: "/deals", section: "trading",
				needs: []uint{model.MgrRightAccRead, model.MgrRightTradesRead}},
		}},

	{key: "dealing", label: "Dealing", icon: "gavel", route: "/dealing", section: "dealing",
		needs: []uint{model.MgrRightTradesDealer}, counter: "dealing"},

	{key: "exposure", label: "Exposure", icon: "scale", route: "/exposure", section: "dealing",
		needs: []uint{model.MgrRightRiskManager}},

	{key: "clients", label: "Clients", icon: "briefcase", route: "/clients", section: "backoffice",
		needs: []uint{model.MgrRightClientsAccess}},

	{key: "payments", label: "Payments", icon: "credit-card", route: "", section: "payments",
		needs: []uint{model.MgrRightCfgPayments},
		children: []navNode{
			{key: "payments_processing", label: "Processing Payments", icon: "hourglass",
				route: "/payments/processing", section: "payments", needs: []uint{model.MgrRightCfgPayments}},
			{key: "payments_active", label: "Active Payments", icon: "activity",
				route: "/payments/active", section: "payments", needs: []uint{model.MgrRightCfgPayments}},
			{key: "payments_history", label: "History of Payments", icon: "archive",
				route: "/payments/history", section: "payments", needs: []uint{model.MgrRightCfgPayments}},
		}},

	{key: "groups", label: "Groups", icon: "folder-tree", route: "/groups", section: "config",
		needs: []uint{model.MgrRightCfgGroups}, counter: "groups"},

	{key: "config", label: "Configuration", icon: "settings", route: "", section: "config",
		children: []navNode{
			{key: "symbols", label: "Symbols", icon: "tag", route: "/symbols", section: "config",
				needs: []uint{model.MgrRightCfgSymbols}, counter: "symbols"},
			{key: "routing", label: "Request Routing", icon: "git-branch", route: "/routing", section: "config",
				needs: []uint{model.MgrRightCfgRequests}, counter: "routing"},
			{key: "datafeeds", label: "Datafeeds", icon: "rss", route: "/datafeeds", section: "config",
				needs: []uint{model.MgrRightCfgDatafeeds}},
			{key: "gateways", label: "Gateways", icon: "plug", route: "/gateways", section: "config",
				needs: []uint{model.MgrRightCfgGateways}},
			{key: "managers", label: "Managers", icon: "shield", route: "/managers", section: "config",
				needs: []uint{model.MgrRightCfgManagers}},
			{key: "holidays", label: "Holidays", icon: "calendar", route: "/holidays", section: "config",
				needs: []uint{model.MgrRightCfgHolidays}},
			{key: "server_time", label: "Server Time", icon: "clock", route: "/server-time", section: "config",
				needs: []uint{model.MgrRightCfgTime}},
			{key: "leverages", label: "Leverages", icon: "percent", route: "/leverages", section: "config",
				needs: []uint{model.MgrRightCfgGroups}},
			{key: "allocations", label: "Allocations", icon: "shuffle", route: "/allocations", section: "config",
				needs: []uint{model.MgrRightCfgAllocations}},
			{key: "automations", label: "Automations", icon: "zap", route: "/automations", section: "config",
				needs: []uint{model.MgrRightCfgAutomations}},
			{key: "reports_cfg", label: "Report Settings", icon: "sliders", route: "/config/reports", section: "config",
				needs: []uint{model.MgrRightCfgReports}},
		}},

	{key: "mailbox", label: "Mailbox", icon: "mail", route: "/mailbox", section: "support",
		needs: []uint{model.MgrRightEmail}},

	{key: "news", label: "News", icon: "newspaper", route: "/news", section: "support",
		needs: []uint{model.MgrRightNews}},

	{key: "journal", label: "Journal", icon: "scroll", route: "/journal", section: "support",
		needs: []uint{model.MgrRightSrvJournals}},
}

// GetNavigation is the back office navigator for the calling manager.
//
//	@Id			GetNavigation
//	@Tags		Navigation
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewNavigation}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/navigation [get]
func (s *HttpServer) GetNavigation(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rights := snap.ManagerRights

	counts := s.navCounts(c.UserContext(), snap.IsManager, snap.ManagerGroups, rights)

	out := &ViewNavigation{
		Login:    snap.Login,
		Terminal: terminalOf(snap.ConnectionType),
		Groups:   snap.ManagerGroups,
		Nodes:    permitted(navTree, rights, counts),
		Can:      rights.Flags(),
		Rights:   rights,
	}

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
		out["accounts"] = s.countOf(ctx,
			`SELECT count(*) FROM hst.users u WHERE `+where, args)

		if r.Has(model.MgrRightTradesRead) {
			out["positions"] = s.countOf(ctx,
				`SELECT count(*) FROM hst.positions p JOIN hst.users u ON u.login = p.login WHERE `+where, args)
			out["orders"] = s.countOf(ctx,
				`SELECT count(*) FROM hst.orders o JOIN hst.users u ON u.login = o.login
				  WHERE o.state IN (0, 1, 3, 7, 8, 9) AND `+where, args)
		}
	}

	if r.Has(model.MgrRightCfgGroups) {
		out["groups"] = s.countOf(ctx, `SELECT count(*) FROM hst.groups`, nil)
	}
	if r.Has(model.MgrRightCfgSymbols) {
		out["symbols"] = s.countOf(ctx, `SELECT count(*) FROM hst.symbols`, nil)
	}
	if r.Has(model.MgrRightCfgRequests) {
		out["routing"] = s.countOf(ctx, `SELECT count(*) FROM hst.routing`, nil)
	}

	return out
}

// countOf answers zero rather than failing the navigator on one bad count.
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

// terminalOf names the panel a connection type belongs to.
func terminalOf(t int32) string {
	if model.UsersConnectionTypes(t) >= 32 && model.UsersConnectionTypes(t) <= 33 {
		if model.UsersConnectionTypes(t) == 32 {
			return "administrator"
		}
		return "manager"
	}

	return "manager"
}
