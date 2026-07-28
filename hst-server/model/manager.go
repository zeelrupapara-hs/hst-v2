package model

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

// Manager is the back office capability attached to a login. Its presence is
// what makes a login staff. Every right is 1 granted or 0 not granted.
type Manager struct {
	Login               int64        `db:"login" json:"login"`
	Name                string       `db:"name" json:"name"`
	Mailbox             string       `db:"mailbox" json:"mailbox"`
	Server              int32        `db:"server" json:"server"`
	RequestLimitLogs    ManagerLimit `db:"request_limit_logs" json:"request_limit_logs"`
	RequestLimitReports ManagerLimit `db:"request_limit_reports" json:"request_limit_reports"`
	Groups              []string     `db:"groups" json:"groups"`

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

// PermitsTerminal reports whether this manager may connect with the given
// terminal type. MT5 gates the admin and manager terminals separately.
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

// CanDeal reports whether the manager may work the dealing desk.
// MT5 requires trades_read before trades_dealer takes effect.
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
