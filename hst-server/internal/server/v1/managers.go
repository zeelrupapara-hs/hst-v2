package v1

import (
	"context"
	"errors"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/pkg/oauth2"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CrtManager promotes an existing login to staff.
type CrtManager struct {
	Login               int64              `json:"login" validate:"required,gt=0"`
	Name                string             `json:"name" validate:"required,max=128"`
	Mailbox             string             `json:"mailbox" validate:"max=128"`
	Server              int32              `json:"server"`
	RequestLimitLogs    model.ManagerLimit `json:"request_limit_logs" validate:"gte=0,lte=6"`
	RequestLimitReports model.ManagerLimit `json:"request_limit_reports" validate:"gte=0,lte=6"`
	Groups              []string           `json:"groups"`
	// Rights are column names from ManagerRightsNames; anything else is rejected
	Rights []string `json:"rights"`
}

// UptManager rewrites the staff role of a login.
type UptManager struct {
	Name                string             `json:"name" validate:"required,max=128"`
	Mailbox             string             `json:"mailbox" validate:"max=128"`
	Server              int32              `json:"server"`
	RequestLimitLogs    model.ManagerLimit `json:"request_limit_logs" validate:"gte=0,lte=6"`
	RequestLimitReports model.ManagerLimit `json:"request_limit_reports" validate:"gte=0,lte=6"`
	Groups              []string           `json:"groups"`
	Rights              []string           `json:"rights"`
}

// ViewManagerRights is the decoded right set, for the admin UI.
type ViewManagerRights struct {
	Login  int64           `json:"login"`
	Name   string          `json:"name"`
	Groups []string        `json:"groups"`
	Rights map[string]bool `json:"rights"`
}

// GetManager returns the whole manager record.
//
//	@Id			GetManager
//	@Tags		Managers
//	@Produce	json
//	@Success	200	{object}	Response{data=model.Manager}
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers/{login} [get]
func (s *HttpServer) GetManager(c *fiber.Ctx) error {
	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	m, err := s.selectManager(c.UserContext(), int64(login))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, m)
}

// GetManagerRights returns the 77 flags of one manager.
//
//	@Id			GetManagerRights
//	@Tags		Managers
//	@Produce	json
//	@Success	200	{object}	Response{data=ViewManagerRights}
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers/{login}/rights [get]
func (s *HttpServer) GetManagerRights(c *fiber.Ctx) error {
	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	m, err := s.selectManager(c.UserContext(), int64(login))
	if err == pgx.ErrNoRows {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, &ViewManagerRights{
		Login:  m.Login,
		Name:   m.Name,
		Groups: m.Groups,
		Rights: oauth2.PackManagerRights(m).Flags(),
	})
}

// ListManagers returns the managers, newest first.
//
//	@Id			ListManagers
//	@Tags		Managers
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewManagerRights}
//	@Failure	400	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers [get]
func (s *HttpServer) ListManagers(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, utils.NewSortable("login", "name"), "login")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT m.login, m.name, m.groups
		   FROM hst.managers m
		  ORDER BY m.login
		  LIMIT $1 OFFSET $2`, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewManagerRights{}
	for rows.Next() {
		var v ViewManagerRights
		if err := rows.Scan(&v.Login, &v.Name, &v.Groups); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// selectManager reads the whole right set in one round trip.
func (s *HttpServer) selectManager(ctx context.Context, login int64) (*model.Manager, error) {
	m := &model.Manager{}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT login, name, mailbox, server, request_limit_logs,
		        request_limit_reports, groups, access,
		        right_admin, right_manager, right_cfg_time, right_cfg_holidays,
		        right_cfg_groups, right_cfg_managers, right_cfg_requests,
		        right_cfg_gateways, right_cfg_datafeeds, right_cfg_reports,
		        right_cfg_symbols, right_cfg_web_services, right_cfg_messengers,
		        right_cfg_kyc, right_cfg_automations, right_cfg_allocations,
		        right_cfg_corporate, right_cfg_payments, right_cfg_mails,
		        right_cfg_streaming, right_srv_journals, right_srv_reports,
		        right_charts, right_email, right_news, right_export,
		        right_techsupport, right_market, right_accountant, right_acc_read,
		        right_acc_details_name, right_acc_details_location,
		        right_acc_details_address, right_acc_details_id,
		        right_acc_details_email, right_acc_details_phone,
		        right_acc_details_general, right_acc_technical,
		        right_acc_tech_modify, right_acc_manager, right_acc_delete,
		        right_acc_online, right_confirm_actions, right_notifications,
		        right_trades_read, right_trades_manager, right_trades_delete,
		        right_trades_dealer, right_trades_supervisor, right_quotes_raw,
		        right_quotes, right_symbol_details, right_risk_manager,
		        right_group_margin, right_group_commission, right_reports,
		        right_clients_access, right_clients_create, right_clients_edit,
		        right_clients_delete, right_clients_kyc,
		        right_clients_details_name, right_clients_details_location,
		        right_clients_details_address, right_clients_details_id,
		        right_clients_details_email, right_clients_details_phone,
		        right_clients_details_general, right_documents_access,
		        right_documents_create, right_documents_edit,
		        right_documents_delete, right_documents_files_add,
		        right_documents_files_delete, right_comments_access,
		        right_comments_create, right_comments_delete
		   FROM hst.managers WHERE login = $1`, login).
		Scan(&m.Login, &m.Name, &m.Mailbox, &m.Server, &m.RequestLimitLogs,
			&m.RequestLimitReports, &m.Groups, &m.Access,
			&m.RightAdmin,
			&m.RightManager,
			&m.RightCfgTime,
			&m.RightCfgHolidays,
			&m.RightCfgGroups,
			&m.RightCfgManagers,
			&m.RightCfgRequests,
			&m.RightCfgGateways,
			&m.RightCfgDatafeeds,
			&m.RightCfgReports,
			&m.RightCfgSymbols,
			&m.RightCfgWebServices,
			&m.RightCfgMessengers,
			&m.RightCfgKyc,
			&m.RightCfgAutomations,
			&m.RightCfgAllocations,
			&m.RightCfgCorporate,
			&m.RightCfgPayments,
			&m.RightCfgMails,
			&m.RightCfgStreaming,
			&m.RightSrvJournals,
			&m.RightSrvReports,
			&m.RightCharts,
			&m.RightEmail,
			&m.RightNews,
			&m.RightExport,
			&m.RightTechsupport,
			&m.RightMarket,
			&m.RightAccountant,
			&m.RightAccRead,
			&m.RightAccDetailsName,
			&m.RightAccDetailsLocation,
			&m.RightAccDetailsAddress,
			&m.RightAccDetailsId,
			&m.RightAccDetailsEmail,
			&m.RightAccDetailsPhone,
			&m.RightAccDetailsGeneral,
			&m.RightAccTechnical,
			&m.RightAccTechModify,
			&m.RightAccManager,
			&m.RightAccDelete,
			&m.RightAccOnline,
			&m.RightConfirmActions,
			&m.RightNotifications,
			&m.RightTradesRead,
			&m.RightTradesManager,
			&m.RightTradesDelete,
			&m.RightTradesDealer,
			&m.RightTradesSupervisor,
			&m.RightQuotesRaw,
			&m.RightQuotes,
			&m.RightSymbolDetails,
			&m.RightRiskManager,
			&m.RightGroupMargin,
			&m.RightGroupCommission,
			&m.RightReports,
			&m.RightClientsAccess,
			&m.RightClientsCreate,
			&m.RightClientsEdit,
			&m.RightClientsDelete,
			&m.RightClientsKyc,
			&m.RightClientsDetailsName,
			&m.RightClientsDetailsLocation,
			&m.RightClientsDetailsAddress,
			&m.RightClientsDetailsId,
			&m.RightClientsDetailsEmail,
			&m.RightClientsDetailsPhone,
			&m.RightClientsDetailsGeneral,
			&m.RightDocumentsAccess,
			&m.RightDocumentsCreate,
			&m.RightDocumentsEdit,
			&m.RightDocumentsDelete,
			&m.RightDocumentsFilesAdd,
			&m.RightDocumentsFilesDelete,
			&m.RightCommentsAccess,
			&m.RightCommentsCreate,
			&m.RightCommentsDelete)

	return m, err
}

// managerRightColumns is every right column, in the order the migration declares them.
const managerRightColumns = `
		right_admin, right_manager, right_cfg_time, right_cfg_holidays,
		right_cfg_groups, right_cfg_managers, right_cfg_requests,
		right_cfg_gateways, right_cfg_datafeeds, right_cfg_reports,
		right_cfg_symbols, right_cfg_web_services, right_cfg_messengers,
		right_cfg_kyc, right_cfg_automations, right_cfg_allocations,
		right_cfg_corporate, right_cfg_payments, right_cfg_mails,
		right_cfg_streaming, right_srv_journals, right_srv_reports, right_charts,
		right_email, right_news, right_export, right_techsupport, right_market,
		right_accountant, right_acc_read, right_acc_details_name,
		right_acc_details_location, right_acc_details_address,
		right_acc_details_id, right_acc_details_email, right_acc_details_phone,
		right_acc_details_general, right_acc_technical, right_acc_tech_modify,
		right_acc_manager, right_acc_delete, right_acc_online,
		right_confirm_actions, right_notifications, right_trades_read,
		right_trades_manager, right_trades_delete, right_trades_dealer,
		right_trades_supervisor, right_quotes_raw, right_quotes,
		right_symbol_details, right_risk_manager, right_group_margin,
		right_group_commission, right_reports, right_clients_access,
		right_clients_create, right_clients_edit, right_clients_delete,
		right_clients_kyc, right_clients_details_name,
		right_clients_details_location, right_clients_details_address,
		right_clients_details_id, right_clients_details_email,
		right_clients_details_phone, right_clients_details_general,
		right_documents_access, right_documents_create, right_documents_edit,
		right_documents_delete, right_documents_files_add,
		right_documents_files_delete, right_comments_access, right_comments_create,
		right_comments_delete`

// b renders a right as the 1 or 0 the schema stores.
func b(granted bool) int32 {
	if granted {
		return 1
	}
	return 0
}

// insertManager attaches the staff role to an existing login.
func insertManager(ctx context.Context, q *pgxpool.Pool, login int64, m *CrtManager,
	r model.ManagerRights, now int64) error {

	// groups is NOT NULL, and an omitted json array arrives as nil.
	groups := m.Groups
	if groups == nil {
		groups = []string{}
	}

	_, err := q.Exec(ctx,
		`INSERT INTO hst.managers (login, name, mailbox, server, request_limit_logs,
		    request_limit_reports, groups, updated_at, `+managerRightColumns+`)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
		$9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25,
		$26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42,
		$43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59,
		$60, $61, $62, $63, $64, $65, $66, $67, $68, $69, $70, $71, $72, $73, $74, $75, $76,
		$77, $78, $79, $80, $81, $82, $83, $84, $85)`,
		login, m.Name, m.Mailbox, m.Server, m.RequestLimitLogs,
		m.RequestLimitReports, groups, now,
		b(r.Has(model.MgrRightAdmin)), b(r.Has(model.MgrRightManager)),
		b(r.Has(model.MgrRightCfgTime)), b(r.Has(model.MgrRightCfgHolidays)),
		b(r.Has(model.MgrRightCfgGroups)), b(r.Has(model.MgrRightCfgManagers)),
		b(r.Has(model.MgrRightCfgRequests)), b(r.Has(model.MgrRightCfgGateways)),
		b(r.Has(model.MgrRightCfgDatafeeds)), b(r.Has(model.MgrRightCfgReports)),
		b(r.Has(model.MgrRightCfgSymbols)), b(r.Has(model.MgrRightCfgWebServices)),
		b(r.Has(model.MgrRightCfgMessengers)), b(r.Has(model.MgrRightCfgKyc)),
		b(r.Has(model.MgrRightCfgAutomations)),
		b(r.Has(model.MgrRightCfgAllocations)), b(r.Has(model.MgrRightCfgCorporate)),
		b(r.Has(model.MgrRightCfgPayments)), b(r.Has(model.MgrRightCfgMails)),
		b(r.Has(model.MgrRightCfgStreaming)), b(r.Has(model.MgrRightSrvJournals)),
		b(r.Has(model.MgrRightSrvReports)), b(r.Has(model.MgrRightCharts)),
		b(r.Has(model.MgrRightEmail)), b(r.Has(model.MgrRightNews)),
		b(r.Has(model.MgrRightExport)), b(r.Has(model.MgrRightTechsupport)),
		b(r.Has(model.MgrRightMarket)), b(r.Has(model.MgrRightAccountant)),
		b(r.Has(model.MgrRightAccRead)), b(r.Has(model.MgrRightAccDetailsName)),
		b(r.Has(model.MgrRightAccDetailsLocation)),
		b(r.Has(model.MgrRightAccDetailsAddress)),
		b(r.Has(model.MgrRightAccDetailsId)),
		b(r.Has(model.MgrRightAccDetailsEmail)),
		b(r.Has(model.MgrRightAccDetailsPhone)),
		b(r.Has(model.MgrRightAccDetailsGeneral)),
		b(r.Has(model.MgrRightAccTechnical)), b(r.Has(model.MgrRightAccTechModify)),
		b(r.Has(model.MgrRightAccManager)), b(r.Has(model.MgrRightAccDelete)),
		b(r.Has(model.MgrRightAccOnline)), b(r.Has(model.MgrRightConfirmActions)),
		b(r.Has(model.MgrRightNotifications)), b(r.Has(model.MgrRightTradesRead)),
		b(r.Has(model.MgrRightTradesManager)), b(r.Has(model.MgrRightTradesDelete)),
		b(r.Has(model.MgrRightTradesDealer)),
		b(r.Has(model.MgrRightTradesSupervisor)), b(r.Has(model.MgrRightQuotesRaw)),
		b(r.Has(model.MgrRightQuotes)), b(r.Has(model.MgrRightSymbolDetails)),
		b(r.Has(model.MgrRightRiskManager)), b(r.Has(model.MgrRightGroupMargin)),
		b(r.Has(model.MgrRightGroupCommission)), b(r.Has(model.MgrRightReports)),
		b(r.Has(model.MgrRightClientsAccess)), b(r.Has(model.MgrRightClientsCreate)),
		b(r.Has(model.MgrRightClientsEdit)), b(r.Has(model.MgrRightClientsDelete)),
		b(r.Has(model.MgrRightClientsKyc)),
		b(r.Has(model.MgrRightClientsDetailsName)),
		b(r.Has(model.MgrRightClientsDetailsLocation)),
		b(r.Has(model.MgrRightClientsDetailsAddress)),
		b(r.Has(model.MgrRightClientsDetailsId)),
		b(r.Has(model.MgrRightClientsDetailsEmail)),
		b(r.Has(model.MgrRightClientsDetailsPhone)),
		b(r.Has(model.MgrRightClientsDetailsGeneral)),
		b(r.Has(model.MgrRightDocumentsAccess)),
		b(r.Has(model.MgrRightDocumentsCreate)),
		b(r.Has(model.MgrRightDocumentsEdit)),
		b(r.Has(model.MgrRightDocumentsDelete)),
		b(r.Has(model.MgrRightDocumentsFilesAdd)),
		b(r.Has(model.MgrRightDocumentsFilesDelete)),
		b(r.Has(model.MgrRightCommentsAccess)),
		b(r.Has(model.MgrRightCommentsCreate)),
		b(r.Has(model.MgrRightCommentsDelete)))

	return err
}

// updateManager rewrites the whole record.
func updateManager(ctx context.Context, q *pgxpool.Pool, login int64, m *UptManager,
	r model.ManagerRights, now int64) (int64, error) {

	groups := m.Groups
	if groups == nil {
		groups = []string{}
	}

	tag, err := q.Exec(ctx,
		`UPDATE hst.managers SET name = $2, mailbox = $3, server = $4,
		    request_limit_logs = $5, request_limit_reports = $6, groups = $7,
		    updated_at = $8,
		    right_admin = $9, right_manager = $10, right_cfg_time = $11,
		    right_cfg_holidays = $12, right_cfg_groups = $13, right_cfg_managers = $14,
		    right_cfg_requests = $15, right_cfg_gateways = $16, right_cfg_datafeeds = $17,
		    right_cfg_reports = $18, right_cfg_symbols = $19, right_cfg_web_services = $20,
		    right_cfg_messengers = $21, right_cfg_kyc = $22, right_cfg_automations = $23,
		    right_cfg_allocations = $24, right_cfg_corporate = $25, right_cfg_payments = $26,
		    right_cfg_mails = $27, right_cfg_streaming = $28, right_srv_journals = $29,
		    right_srv_reports = $30, right_charts = $31, right_email = $32, right_news = $33,
		    right_export = $34, right_techsupport = $35, right_market = $36,
		    right_accountant = $37, right_acc_read = $38, right_acc_details_name = $39,
		    right_acc_details_location = $40, right_acc_details_address = $41,
		    right_acc_details_id = $42, right_acc_details_email = $43,
		    right_acc_details_phone = $44, right_acc_details_general = $45,
		    right_acc_technical = $46, right_acc_tech_modify = $47, right_acc_manager = $48,
		    right_acc_delete = $49, right_acc_online = $50, right_confirm_actions = $51,
		    right_notifications = $52, right_trades_read = $53, right_trades_manager = $54,
		    right_trades_delete = $55, right_trades_dealer = $56,
		    right_trades_supervisor = $57, right_quotes_raw = $58, right_quotes = $59,
		    right_symbol_details = $60, right_risk_manager = $61, right_group_margin = $62,
		    right_group_commission = $63, right_reports = $64, right_clients_access = $65,
		    right_clients_create = $66, right_clients_edit = $67, right_clients_delete = $68,
		    right_clients_kyc = $69, right_clients_details_name = $70,
		    right_clients_details_location = $71, right_clients_details_address = $72,
		    right_clients_details_id = $73, right_clients_details_email = $74,
		    right_clients_details_phone = $75, right_clients_details_general = $76,
		    right_documents_access = $77, right_documents_create = $78,
		    right_documents_edit = $79, right_documents_delete = $80,
		    right_documents_files_add = $81, right_documents_files_delete = $82,
		    right_comments_access = $83, right_comments_create = $84,
		    right_comments_delete = $85
		  WHERE login = $1`,
		login, m.Name, m.Mailbox, m.Server, m.RequestLimitLogs,
		m.RequestLimitReports, groups, now,
		b(r.Has(model.MgrRightAdmin)), b(r.Has(model.MgrRightManager)),
		b(r.Has(model.MgrRightCfgTime)), b(r.Has(model.MgrRightCfgHolidays)),
		b(r.Has(model.MgrRightCfgGroups)), b(r.Has(model.MgrRightCfgManagers)),
		b(r.Has(model.MgrRightCfgRequests)), b(r.Has(model.MgrRightCfgGateways)),
		b(r.Has(model.MgrRightCfgDatafeeds)), b(r.Has(model.MgrRightCfgReports)),
		b(r.Has(model.MgrRightCfgSymbols)), b(r.Has(model.MgrRightCfgWebServices)),
		b(r.Has(model.MgrRightCfgMessengers)), b(r.Has(model.MgrRightCfgKyc)),
		b(r.Has(model.MgrRightCfgAutomations)),
		b(r.Has(model.MgrRightCfgAllocations)), b(r.Has(model.MgrRightCfgCorporate)),
		b(r.Has(model.MgrRightCfgPayments)), b(r.Has(model.MgrRightCfgMails)),
		b(r.Has(model.MgrRightCfgStreaming)), b(r.Has(model.MgrRightSrvJournals)),
		b(r.Has(model.MgrRightSrvReports)), b(r.Has(model.MgrRightCharts)),
		b(r.Has(model.MgrRightEmail)), b(r.Has(model.MgrRightNews)),
		b(r.Has(model.MgrRightExport)), b(r.Has(model.MgrRightTechsupport)),
		b(r.Has(model.MgrRightMarket)), b(r.Has(model.MgrRightAccountant)),
		b(r.Has(model.MgrRightAccRead)), b(r.Has(model.MgrRightAccDetailsName)),
		b(r.Has(model.MgrRightAccDetailsLocation)),
		b(r.Has(model.MgrRightAccDetailsAddress)),
		b(r.Has(model.MgrRightAccDetailsId)),
		b(r.Has(model.MgrRightAccDetailsEmail)),
		b(r.Has(model.MgrRightAccDetailsPhone)),
		b(r.Has(model.MgrRightAccDetailsGeneral)),
		b(r.Has(model.MgrRightAccTechnical)), b(r.Has(model.MgrRightAccTechModify)),
		b(r.Has(model.MgrRightAccManager)), b(r.Has(model.MgrRightAccDelete)),
		b(r.Has(model.MgrRightAccOnline)), b(r.Has(model.MgrRightConfirmActions)),
		b(r.Has(model.MgrRightNotifications)), b(r.Has(model.MgrRightTradesRead)),
		b(r.Has(model.MgrRightTradesManager)), b(r.Has(model.MgrRightTradesDelete)),
		b(r.Has(model.MgrRightTradesDealer)),
		b(r.Has(model.MgrRightTradesSupervisor)), b(r.Has(model.MgrRightQuotesRaw)),
		b(r.Has(model.MgrRightQuotes)), b(r.Has(model.MgrRightSymbolDetails)),
		b(r.Has(model.MgrRightRiskManager)), b(r.Has(model.MgrRightGroupMargin)),
		b(r.Has(model.MgrRightGroupCommission)), b(r.Has(model.MgrRightReports)),
		b(r.Has(model.MgrRightClientsAccess)), b(r.Has(model.MgrRightClientsCreate)),
		b(r.Has(model.MgrRightClientsEdit)), b(r.Has(model.MgrRightClientsDelete)),
		b(r.Has(model.MgrRightClientsKyc)),
		b(r.Has(model.MgrRightClientsDetailsName)),
		b(r.Has(model.MgrRightClientsDetailsLocation)),
		b(r.Has(model.MgrRightClientsDetailsAddress)),
		b(r.Has(model.MgrRightClientsDetailsId)),
		b(r.Has(model.MgrRightClientsDetailsEmail)),
		b(r.Has(model.MgrRightClientsDetailsPhone)),
		b(r.Has(model.MgrRightClientsDetailsGeneral)),
		b(r.Has(model.MgrRightDocumentsAccess)),
		b(r.Has(model.MgrRightDocumentsCreate)),
		b(r.Has(model.MgrRightDocumentsEdit)),
		b(r.Has(model.MgrRightDocumentsDelete)),
		b(r.Has(model.MgrRightDocumentsFilesAdd)),
		b(r.Has(model.MgrRightDocumentsFilesDelete)),
		b(r.Has(model.MgrRightCommentsAccess)),
		b(r.Has(model.MgrRightCommentsCreate)),
		b(r.Has(model.MgrRightCommentsDelete)))
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// packRightNames turns right names into the bitset, rejecting unknown names.
func packRightNames(names []string) (model.ManagerRights, bool) {
	byName := make(map[string]uint, model.ManagerRightsCount)
	for bit, name := range model.ManagerRightsNames {
		byName[name] = bit
	}

	var r model.ManagerRights
	for _, n := range names {
		bit, ok := byName[n]
		if !ok {
			return r, false
		}
		r = r.Set(bit)
	}

	return r, true
}

// CreateManager promotes an existing login to staff by attaching a manager row.
//
//	@Id			CreateManager
//	@Tags		Managers
//	@Accept		json
//	@Produce	json
//	@Success	201	{object}	Response{data=model.Manager}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	409	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers [post]
func (s *HttpServer) CreateManager(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtManager
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	rights, ok := packRightNames(body.Rights)
	if !ok {
		return s.App.HttpResponseBadRequest(c, errs.ErrUnknownManagerRight)
	}

	// the login must already exist; a manager is a role on a user, not a user
	var exists bool
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM hst.users WHERE login = $1)`, body.Login).Scan(&exists); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !exists {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	var already bool
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM hst.managers WHERE login = $1)`, body.Login).Scan(&already); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if already {
		return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
	}

	if err := insertManager(ctx, s.DB.DB, body.Login, &body, rights, time.Now().UnixNano()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// the login just became staff, so any session it already holds is stale
	if err := s.OAuth2.InvalidateLogin(ctx, body.Login, model.SessionRevokedRightsChanged); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "manager created",
		"actor", snap.Login, "target", body.Login)

	return s.getManager(c, body.Login, s.App.HttpResponseCreated)
}

// UpdateManager replaces the name, groups and the whole right set of a manager.
//
//	@Id			UpdateManager
//	@Tags		Managers
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	Response{data=model.Manager}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers/{login} [patch]
func (s *HttpServer) UpdateManager(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptManager
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	rights, ok := packRightNames(body.Rights)
	if !ok {
		return s.App.HttpResponseBadRequest(c, errs.ErrUnknownManagerRight)
	}

	affected, err := updateManager(ctx, s.DB.DB, int64(login), &body, rights,
		time.Now().UnixNano())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if affected == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	// rights changed, so every live session of this login must be rebuilt
	if err := s.OAuth2.InvalidateLogin(ctx, int64(login), model.SessionRevokedRightsChanged); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "manager updated",
		"actor", snap.Login, "target", login)

	return s.getManager(c, int64(login), s.App.HttpResponseOK)
}

// DeleteManager removes the staff role.
//
//	@Id			DeleteManager
//	@Tags		Managers
//	@Produce	json
//	@Success	204	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/managers/{login} [delete]
func (s *HttpServer) DeleteManager(c *fiber.Ctx) error {
	ctx := c.UserContext()

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tag, err := s.DB.DB.Exec(ctx, `DELETE FROM hst.managers WHERE login = $1`, login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	// this login is no longer staff.
	if err := s.OAuth2.InvalidateLogin(ctx, int64(login), model.SessionRevokedRightsChanged); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "manager removed",
		"actor", snap.Login, "target", login)

	return s.App.HttpResponseNoContent(c)
}

// getManager reads one manager and answers with the given responder.
func (s *HttpServer) getManager(c *fiber.Ctx, login int64,
	respond func(*fiber.Ctx, interface{}) error) error {

	m, err := s.selectManager(c.UserContext(), login)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return respond(c, m)
}
