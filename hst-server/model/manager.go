package model

import "net/netip"

type ManagerLimit int32

const (
	ManagerLimit_all      ManagerLimit = 0
	ManagerLimit_months_1 ManagerLimit = 1
	ManagerLimit_months_3 ManagerLimit = 2
	ManagerLimit_months_6 ManagerLimit = 3
	ManagerLimit_year_1   ManagerLimit = 4
	ManagerLimit_year_2   ManagerLimit = 5
	ManagerLimit_year_3   ManagerLimit = 6
)

// Enum value maps for ManagerLimit.
var (
	ManagerLimit_name = map[int32]string{
		0: "all",
		1: "months_1",
		2: "months_3",
		3: "months_6",
		4: "year_1",
		5: "year_2",
		6: "year_3",
	}
	ManagerLimit_value = map[string]int32{
		"all":      0,
		"months_1": 1,
		"months_3": 2,
		"months_6": 3,
		"year_1":   4,
		"year_2":   5,
		"year_3":   6,
	}
)

// Manager is the back office capability attached to a login.
type Manager struct {
	Login               int64        `db:"login" json:"login"`
	Name                string       `db:"name" json:"name"`
	Mailbox             string       `db:"mailbox" json:"mailbox"`
	Server              int32        `db:"server" json:"server"`
	RequestLimitLogs    ManagerLimit `db:"request_limit_logs" json:"request_limit_logs"`
	RequestLimitReports ManagerLimit `db:"request_limit_reports" json:"request_limit_reports"`
	Groups              []string     `db:"groups" json:"groups"`
	// Access is the IP allowlist, an empty list means any address is allowed.
	Access []netip.Prefix `db:"access" json:"access"`

	RightAdmin                  int32 `db:"right_admin" json:"right_admin"`
	RightManager                int32 `db:"right_manager" json:"right_manager"`
	RightCfgTime                int32 `db:"right_cfg_time" json:"right_cfg_time"`
	RightCfgHolidays            int32 `db:"right_cfg_holidays" json:"right_cfg_holidays"`
	RightCfgGroups              int32 `db:"right_cfg_groups" json:"right_cfg_groups"`
	RightCfgManagers            int32 `db:"right_cfg_managers" json:"right_cfg_managers"`
	RightCfgRequests            int32 `db:"right_cfg_requests" json:"right_cfg_requests"`
	RightCfgGateways            int32 `db:"right_cfg_gateways" json:"right_cfg_gateways"`
	RightCfgDatafeeds           int32 `db:"right_cfg_datafeeds" json:"right_cfg_datafeeds"`
	RightCfgReports             int32 `db:"right_cfg_reports" json:"right_cfg_reports"`
	RightCfgSymbols             int32 `db:"right_cfg_symbols" json:"right_cfg_symbols"`
	RightCfgWebServices         int32 `db:"right_cfg_web_services" json:"right_cfg_web_services"`
	RightCfgMessengers          int32 `db:"right_cfg_messengers" json:"right_cfg_messengers"`
	RightCfgKyc                 int32 `db:"right_cfg_kyc" json:"right_cfg_kyc"`
	RightCfgAutomations         int32 `db:"right_cfg_automations" json:"right_cfg_automations"`
	RightCfgAllocations         int32 `db:"right_cfg_allocations" json:"right_cfg_allocations"`
	RightCfgCorporate           int32 `db:"right_cfg_corporate" json:"right_cfg_corporate"`
	RightCfgPayments            int32 `db:"right_cfg_payments" json:"right_cfg_payments"`
	RightCfgMails               int32 `db:"right_cfg_mails" json:"right_cfg_mails"`
	RightCfgStreaming           int32 `db:"right_cfg_streaming" json:"right_cfg_streaming"`
	RightSrvJournals            int32 `db:"right_srv_journals" json:"right_srv_journals"`
	RightSrvReports             int32 `db:"right_srv_reports" json:"right_srv_reports"`
	RightCharts                 int32 `db:"right_charts" json:"right_charts"`
	RightEmail                  int32 `db:"right_email" json:"right_email"`
	RightNews                   int32 `db:"right_news" json:"right_news"`
	RightExport                 int32 `db:"right_export" json:"right_export"`
	RightTechsupport            int32 `db:"right_techsupport" json:"right_techsupport"`
	RightMarket                 int32 `db:"right_market" json:"right_market"`
	RightAccountant             int32 `db:"right_accountant" json:"right_accountant"`
	RightAccRead                int32 `db:"right_acc_read" json:"right_acc_read"`
	RightAccDetailsName         int32 `db:"right_acc_details_name" json:"right_acc_details_name"`
	RightAccDetailsLocation     int32 `db:"right_acc_details_location" json:"right_acc_details_location"`
	RightAccDetailsAddress      int32 `db:"right_acc_details_address" json:"right_acc_details_address"`
	RightAccDetailsId           int32 `db:"right_acc_details_id" json:"right_acc_details_id"`
	RightAccDetailsEmail        int32 `db:"right_acc_details_email" json:"right_acc_details_email"`
	RightAccDetailsPhone        int32 `db:"right_acc_details_phone" json:"right_acc_details_phone"`
	RightAccDetailsGeneral      int32 `db:"right_acc_details_general" json:"right_acc_details_general"`
	RightAccTechnical           int32 `db:"right_acc_technical" json:"right_acc_technical"`
	RightAccTechModify          int32 `db:"right_acc_tech_modify" json:"right_acc_tech_modify"`
	RightAccManager             int32 `db:"right_acc_manager" json:"right_acc_manager"`
	RightAccDelete              int32 `db:"right_acc_delete" json:"right_acc_delete"`
	RightAccOnline              int32 `db:"right_acc_online" json:"right_acc_online"`
	RightConfirmActions         int32 `db:"right_confirm_actions" json:"right_confirm_actions"`
	RightNotifications          int32 `db:"right_notifications" json:"right_notifications"`
	RightTradesRead             int32 `db:"right_trades_read" json:"right_trades_read"`
	RightTradesManager          int32 `db:"right_trades_manager" json:"right_trades_manager"`
	RightTradesDelete           int32 `db:"right_trades_delete" json:"right_trades_delete"`
	RightTradesDealer           int32 `db:"right_trades_dealer" json:"right_trades_dealer"`
	RightTradesSupervisor       int32 `db:"right_trades_supervisor" json:"right_trades_supervisor"`
	RightQuotesRaw              int32 `db:"right_quotes_raw" json:"right_quotes_raw"`
	RightQuotes                 int32 `db:"right_quotes" json:"right_quotes"`
	RightSymbolDetails          int32 `db:"right_symbol_details" json:"right_symbol_details"`
	RightRiskManager            int32 `db:"right_risk_manager" json:"right_risk_manager"`
	RightGroupMargin            int32 `db:"right_group_margin" json:"right_group_margin"`
	RightGroupCommission        int32 `db:"right_group_commission" json:"right_group_commission"`
	RightReports                int32 `db:"right_reports" json:"right_reports"`
	RightClientsAccess          int32 `db:"right_clients_access" json:"right_clients_access"`
	RightClientsCreate          int32 `db:"right_clients_create" json:"right_clients_create"`
	RightClientsEdit            int32 `db:"right_clients_edit" json:"right_clients_edit"`
	RightClientsDelete          int32 `db:"right_clients_delete" json:"right_clients_delete"`
	RightClientsKyc             int32 `db:"right_clients_kyc" json:"right_clients_kyc"`
	RightClientsDetailsName     int32 `db:"right_clients_details_name" json:"right_clients_details_name"`
	RightClientsDetailsLocation int32 `db:"right_clients_details_location" json:"right_clients_details_location"`
	RightClientsDetailsAddress  int32 `db:"right_clients_details_address" json:"right_clients_details_address"`
	RightClientsDetailsId       int32 `db:"right_clients_details_id" json:"right_clients_details_id"`
	RightClientsDetailsEmail    int32 `db:"right_clients_details_email" json:"right_clients_details_email"`
	RightClientsDetailsPhone    int32 `db:"right_clients_details_phone" json:"right_clients_details_phone"`
	RightClientsDetailsGeneral  int32 `db:"right_clients_details_general" json:"right_clients_details_general"`
	RightDocumentsAccess        int32 `db:"right_documents_access" json:"right_documents_access"`
	RightDocumentsCreate        int32 `db:"right_documents_create" json:"right_documents_create"`
	RightDocumentsEdit          int32 `db:"right_documents_edit" json:"right_documents_edit"`
	RightDocumentsDelete        int32 `db:"right_documents_delete" json:"right_documents_delete"`
	RightDocumentsFilesAdd      int32 `db:"right_documents_files_add" json:"right_documents_files_add"`
	RightDocumentsFilesDelete   int32 `db:"right_documents_files_delete" json:"right_documents_files_delete"`
	RightCommentsAccess         int32 `db:"right_comments_access" json:"right_comments_access"`
	RightCommentsCreate         int32 `db:"right_comments_create" json:"right_comments_create"`
	RightCommentsDelete         int32 `db:"right_comments_delete" json:"right_comments_delete"`

	UpdatedAt int64 `db:"updated_at" json:"updated_at"`
}

func (Manager) TableName() string { return "hst.managers" }

// PermitsTerminal reports whether this terminal type is allowed.
func (m *Manager) PermitsTerminal(t UsersConnectionTypes) bool {
	switch t {
	case UsersConnectionTypes_admin, UsersConnectionTypes_admin_api:
		return m.RightAdmin == 1
	case UsersConnectionTypes_manager, UsersConnectionTypes_manager_api,
		UsersConnectionTypes_manager_api_web:
		return m.RightManager == 1
	default:
		return false
	}
}

// PermitsIP reports whether the manager may connect from this address.
func (m *Manager) PermitsIP(ip string) bool {
	if len(m.Access) == 0 {
		return true
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	for _, allowed := range m.Access {
		if allowed.Contains(addr) {
			return true
		}
	}

	return false
}

// CanDeal reports whether the manager may work the dealing desk.
func (m *Manager) CanDeal() bool {
	return m.RightTradesRead == 1 && m.RightTradesDealer == 1
}

// CanManageTrades reports whether the manager may modify orders and positions.
func (m *Manager) CanManageTrades() bool {
	return m.RightTradesRead == 1 && m.RightTradesManager == 1
}

// CanDeleteTrades requires trades_manager per MT5.
func (m *Manager) CanDeleteTrades() bool {
	return m.CanManageTrades() && m.RightTradesDelete == 1
}

// CanDeleteAccounts requires acc_manager per MT5.
func (m *Manager) CanDeleteAccounts() bool {
	return m.RightAccManager == 1 && m.RightAccDelete == 1
}

// ManagerRights packs the 77 right_ columns of hst.managers into two words.
type ManagerRights [2]uint64

// Bit indices follow the column order in the managers migration.
const (
	MgrRightAdmin                  uint = 0
	MgrRightManager                uint = 1
	MgrRightCfgTime                uint = 2
	MgrRightCfgHolidays            uint = 3
	MgrRightCfgGroups              uint = 4
	MgrRightCfgManagers            uint = 5
	MgrRightCfgRequests            uint = 6
	MgrRightCfgGateways            uint = 7
	MgrRightCfgDatafeeds           uint = 8
	MgrRightCfgReports             uint = 9
	MgrRightCfgSymbols             uint = 10
	MgrRightCfgWebServices         uint = 11
	MgrRightCfgMessengers          uint = 12
	MgrRightCfgKyc                 uint = 13
	MgrRightCfgAutomations         uint = 14
	MgrRightCfgAllocations         uint = 15
	MgrRightCfgCorporate           uint = 16
	MgrRightCfgPayments            uint = 17
	MgrRightCfgMails               uint = 18
	MgrRightCfgStreaming           uint = 19
	MgrRightSrvJournals            uint = 20
	MgrRightSrvReports             uint = 21
	MgrRightCharts                 uint = 22
	MgrRightEmail                  uint = 23
	MgrRightNews                   uint = 24
	MgrRightExport                 uint = 25
	MgrRightTechsupport            uint = 26
	MgrRightMarket                 uint = 27
	MgrRightAccountant             uint = 28
	MgrRightAccRead                uint = 29
	MgrRightAccDetailsName         uint = 30
	MgrRightAccDetailsLocation     uint = 31
	MgrRightAccDetailsAddress      uint = 32
	MgrRightAccDetailsId           uint = 33
	MgrRightAccDetailsEmail        uint = 34
	MgrRightAccDetailsPhone        uint = 35
	MgrRightAccDetailsGeneral      uint = 36
	MgrRightAccTechnical           uint = 37
	MgrRightAccTechModify          uint = 38
	MgrRightAccManager             uint = 39
	MgrRightAccDelete              uint = 40
	MgrRightAccOnline              uint = 41
	MgrRightConfirmActions         uint = 42
	MgrRightNotifications          uint = 43
	MgrRightTradesRead             uint = 44
	MgrRightTradesManager          uint = 45
	MgrRightTradesDelete           uint = 46
	MgrRightTradesDealer           uint = 47
	MgrRightTradesSupervisor       uint = 48
	MgrRightQuotesRaw              uint = 49
	MgrRightQuotes                 uint = 50
	MgrRightSymbolDetails          uint = 51
	MgrRightRiskManager            uint = 52
	MgrRightGroupMargin            uint = 53
	MgrRightGroupCommission        uint = 54
	MgrRightReports                uint = 55
	MgrRightClientsAccess          uint = 56
	MgrRightClientsCreate          uint = 57
	MgrRightClientsEdit            uint = 58
	MgrRightClientsDelete          uint = 59
	MgrRightClientsKyc             uint = 60
	MgrRightClientsDetailsName     uint = 61
	MgrRightClientsDetailsLocation uint = 62
	MgrRightClientsDetailsAddress  uint = 63
	MgrRightClientsDetailsId       uint = 64
	MgrRightClientsDetailsEmail    uint = 65
	MgrRightClientsDetailsPhone    uint = 66
	MgrRightClientsDetailsGeneral  uint = 67
	MgrRightDocumentsAccess        uint = 68
	MgrRightDocumentsCreate        uint = 69
	MgrRightDocumentsEdit          uint = 70
	MgrRightDocumentsDelete        uint = 71
	MgrRightDocumentsFilesAdd      uint = 72
	MgrRightDocumentsFilesDelete   uint = 73
	MgrRightCommentsAccess         uint = 74
	MgrRightCommentsCreate         uint = 75
	MgrRightCommentsDelete         uint = 76
)

// ManagerRightsCount is how many bits are in use.
const ManagerRightsCount = 77

// ManagerRightsNames maps each bit to its column name, for the admin UI.
var ManagerRightsNames = map[uint]string{
	MgrRightAdmin:                  "right_admin",
	MgrRightManager:                "right_manager",
	MgrRightCfgTime:                "right_cfg_time",
	MgrRightCfgHolidays:            "right_cfg_holidays",
	MgrRightCfgGroups:              "right_cfg_groups",
	MgrRightCfgManagers:            "right_cfg_managers",
	MgrRightCfgRequests:            "right_cfg_requests",
	MgrRightCfgGateways:            "right_cfg_gateways",
	MgrRightCfgDatafeeds:           "right_cfg_datafeeds",
	MgrRightCfgReports:             "right_cfg_reports",
	MgrRightCfgSymbols:             "right_cfg_symbols",
	MgrRightCfgWebServices:         "right_cfg_web_services",
	MgrRightCfgMessengers:          "right_cfg_messengers",
	MgrRightCfgKyc:                 "right_cfg_kyc",
	MgrRightCfgAutomations:         "right_cfg_automations",
	MgrRightCfgAllocations:         "right_cfg_allocations",
	MgrRightCfgCorporate:           "right_cfg_corporate",
	MgrRightCfgPayments:            "right_cfg_payments",
	MgrRightCfgMails:               "right_cfg_mails",
	MgrRightCfgStreaming:           "right_cfg_streaming",
	MgrRightSrvJournals:            "right_srv_journals",
	MgrRightSrvReports:             "right_srv_reports",
	MgrRightCharts:                 "right_charts",
	MgrRightEmail:                  "right_email",
	MgrRightNews:                   "right_news",
	MgrRightExport:                 "right_export",
	MgrRightTechsupport:            "right_techsupport",
	MgrRightMarket:                 "right_market",
	MgrRightAccountant:             "right_accountant",
	MgrRightAccRead:                "right_acc_read",
	MgrRightAccDetailsName:         "right_acc_details_name",
	MgrRightAccDetailsLocation:     "right_acc_details_location",
	MgrRightAccDetailsAddress:      "right_acc_details_address",
	MgrRightAccDetailsId:           "right_acc_details_id",
	MgrRightAccDetailsEmail:        "right_acc_details_email",
	MgrRightAccDetailsPhone:        "right_acc_details_phone",
	MgrRightAccDetailsGeneral:      "right_acc_details_general",
	MgrRightAccTechnical:           "right_acc_technical",
	MgrRightAccTechModify:          "right_acc_tech_modify",
	MgrRightAccManager:             "right_acc_manager",
	MgrRightAccDelete:              "right_acc_delete",
	MgrRightAccOnline:              "right_acc_online",
	MgrRightConfirmActions:         "right_confirm_actions",
	MgrRightNotifications:          "right_notifications",
	MgrRightTradesRead:             "right_trades_read",
	MgrRightTradesManager:          "right_trades_manager",
	MgrRightTradesDelete:           "right_trades_delete",
	MgrRightTradesDealer:           "right_trades_dealer",
	MgrRightTradesSupervisor:       "right_trades_supervisor",
	MgrRightQuotesRaw:              "right_quotes_raw",
	MgrRightQuotes:                 "right_quotes",
	MgrRightSymbolDetails:          "right_symbol_details",
	MgrRightRiskManager:            "right_risk_manager",
	MgrRightGroupMargin:            "right_group_margin",
	MgrRightGroupCommission:        "right_group_commission",
	MgrRightReports:                "right_reports",
	MgrRightClientsAccess:          "right_clients_access",
	MgrRightClientsCreate:          "right_clients_create",
	MgrRightClientsEdit:            "right_clients_edit",
	MgrRightClientsDelete:          "right_clients_delete",
	MgrRightClientsKyc:             "right_clients_kyc",
	MgrRightClientsDetailsName:     "right_clients_details_name",
	MgrRightClientsDetailsLocation: "right_clients_details_location",
	MgrRightClientsDetailsAddress:  "right_clients_details_address",
	MgrRightClientsDetailsId:       "right_clients_details_id",
	MgrRightClientsDetailsEmail:    "right_clients_details_email",
	MgrRightClientsDetailsPhone:    "right_clients_details_phone",
	MgrRightClientsDetailsGeneral:  "right_clients_details_general",
	MgrRightDocumentsAccess:        "right_documents_access",
	MgrRightDocumentsCreate:        "right_documents_create",
	MgrRightDocumentsEdit:          "right_documents_edit",
	MgrRightDocumentsDelete:        "right_documents_delete",
	MgrRightDocumentsFilesAdd:      "right_documents_files_add",
	MgrRightDocumentsFilesDelete:   "right_documents_files_delete",
	MgrRightCommentsAccess:         "right_comments_access",
	MgrRightCommentsCreate:         "right_comments_create",
	MgrRightCommentsDelete:         "right_comments_delete",
}

// Has reports whether the right is granted.
func (r ManagerRights) Has(bit uint) bool {
	if bit >= ManagerRightsCount {
		return false
	}
	return r[bit>>6]&(1<<(bit&63)) != 0
}

// Set returns r with the right granted.
func (r ManagerRights) Set(bit uint) ManagerRights {
	if bit < ManagerRightsCount {
		r[bit>>6] |= 1 << (bit & 63)
	}
	return r
}

// PermitsTerminal reports whether the manager may connect with this terminal type.
func (r ManagerRights) PermitsTerminal(t UsersConnectionTypes) bool {
	switch t {
	case UsersConnectionTypes_admin, UsersConnectionTypes_admin_api:
		return r.Has(MgrRightAdmin)
	case UsersConnectionTypes_manager, UsersConnectionTypes_manager_api,
		UsersConnectionTypes_manager_api_web:
		return r.Has(MgrRightManager)
	default:
		return false
	}
}

// CanDeal reports whether the manager may work the dealing desk.
func (r ManagerRights) CanDeal() bool {
	return r.Has(MgrRightTradesRead) && r.Has(MgrRightTradesDealer)
}

// CanManageTrades reports whether the manager may modify orders and positions.
func (r ManagerRights) CanManageTrades() bool {
	return r.Has(MgrRightTradesRead) && r.Has(MgrRightTradesManager)
}

// CanDeleteTrades requires trades_manager per MT5.
func (r ManagerRights) CanDeleteTrades() bool {
	return r.CanManageTrades() && r.Has(MgrRightTradesDelete)
}

// CanDeleteAccounts requires acc_manager per MT5.
func (r ManagerRights) CanDeleteAccounts() bool {
	return r.Has(MgrRightAccManager) && r.Has(MgrRightAccDelete)
}

// Flags decodes the bitset back to column name to granted, for the admin UI.
func (r ManagerRights) Flags() map[string]bool {
	out := make(map[string]bool, ManagerRightsCount)
	for bit, name := range ManagerRightsNames {
		out[name] = r.Has(bit)
	}
	return out
}

// ManagerRightsColumns is the column order the login query must select in.
var ManagerRightsColumns = []string{
	"right_admin",
	"right_manager",
	"right_cfg_time",
	"right_cfg_holidays",
	"right_cfg_groups",
	"right_cfg_managers",
	"right_cfg_requests",
	"right_cfg_gateways",
	"right_cfg_datafeeds",
	"right_cfg_reports",
	"right_cfg_symbols",
	"right_cfg_web_services",
	"right_cfg_messengers",
	"right_cfg_kyc",
	"right_cfg_automations",
	"right_cfg_allocations",
	"right_cfg_corporate",
	"right_cfg_payments",
	"right_cfg_mails",
	"right_cfg_streaming",
	"right_srv_journals",
	"right_srv_reports",
	"right_charts",
	"right_email",
	"right_news",
	"right_export",
	"right_techsupport",
	"right_market",
	"right_accountant",
	"right_acc_read",
	"right_acc_details_name",
	"right_acc_details_location",
	"right_acc_details_address",
	"right_acc_details_id",
	"right_acc_details_email",
	"right_acc_details_phone",
	"right_acc_details_general",
	"right_acc_technical",
	"right_acc_tech_modify",
	"right_acc_manager",
	"right_acc_delete",
	"right_acc_online",
	"right_confirm_actions",
	"right_notifications",
	"right_trades_read",
	"right_trades_manager",
	"right_trades_delete",
	"right_trades_dealer",
	"right_trades_supervisor",
	"right_quotes_raw",
	"right_quotes",
	"right_symbol_details",
	"right_risk_manager",
	"right_group_margin",
	"right_group_commission",
	"right_reports",
	"right_clients_access",
	"right_clients_create",
	"right_clients_edit",
	"right_clients_delete",
	"right_clients_kyc",
	"right_clients_details_name",
	"right_clients_details_location",
	"right_clients_details_address",
	"right_clients_details_id",
	"right_clients_details_email",
	"right_clients_details_phone",
	"right_clients_details_general",
	"right_documents_access",
	"right_documents_create",
	"right_documents_edit",
	"right_documents_delete",
	"right_documents_files_add",
	"right_documents_files_delete",
	"right_comments_access",
	"right_comments_create",
	"right_comments_delete",
}
