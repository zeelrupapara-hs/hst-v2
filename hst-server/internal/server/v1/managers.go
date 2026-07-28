package v1

import (
	"context"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/oauth2"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewManagerRights is the decoded right set, for the admin UI.
type ViewManagerRights struct {
	Login  int64           `json:"login"`
	Name   string          `json:"name"`
	Groups []string        `json:"groups"`
	Rights map[string]bool `json:"rights"`
}

// GetManagerRights returns the 77 flags of one manager.
//
// @Id			GetManagerRights
// @Tags		Managers
// @Produce		json
// @Success		200	{object}	ViewManagerRights
// @Security	BearerAuth
// @Router		/api/v1/managers/{login}/rights [get]
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
// @Id			ListManagers
// @Tags		Managers
// @Produce		json
// @Security	BearerAuth
// @Router		/api/v1/managers [get]
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

// selectManager reads the whole right set in one round trip. Login pays this
// once, then every later request answers from the packed bitset.
func (s *HttpServer) selectManager(ctx context.Context, login int64) (*model.Manager, error) {
	m := &model.Manager{}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT login, name, mailbox, server, request_limit_logs,
		        request_limit_reports, groups,
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
			&m.RequestLimitReports, &m.Groups,
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

// b renders a right as the 1 or 0 the schema stores.
func b(granted bool) int32 {
	if granted {
		return 1
	}
	return 0
}

// insertManager writes the manager row inside the caller's transaction, so a
// user without its rights row can never be committed.
func insertManager(ctx context.Context, tx pgx.Tx, login int64, m *CrtManager,
	r model.ManagerRights, now int64) error {

	// groups is NOT NULL, and an omitted json array arrives as nil, which pgx
	// would send as NULL. An empty set means "no group filter", not "unset".
	if m.Groups == nil {
		m.Groups = []string{}
	}

	_, err := tx.Exec(ctx,
		`INSERT INTO hst.managers (login, name, groups, updated_at,
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
		 right_documents_files_delete, right_comments_access,
		 right_comments_create, right_comments_delete)
		 VALUES ($1, $2, $3, $4,
		 $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
		 $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34,
		 $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49,
		 $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64,
		 $65, $66, $67, $68, $69, $70, $71, $72, $73, $74, $75, $76, $77, $78, $79,
		 $80, $81)`,
		login, m.Name, m.Groups, now,
		b(r.Has(model.MgrRightAdmin)), b(r.Has(model.MgrRightManager)),
		b(r.Has(model.MgrRightCfgTime)), b(r.Has(model.MgrRightCfgHolidays)),
		b(r.Has(model.MgrRightCfgGroups)), b(r.Has(model.MgrRightCfgManagers)),
		b(r.Has(model.MgrRightCfgRequests)), b(r.Has(model.MgrRightCfgGateways)),
		b(r.Has(model.MgrRightCfgDatafeeds)), b(r.Has(model.MgrRightCfgReports)),
		b(r.Has(model.MgrRightCfgSymbols)), b(r.Has(model.MgrRightCfgWebServices)),
		b(r.Has(model.MgrRightCfgMessengers)), b(r.Has(model.MgrRightCfgKyc)),
		b(r.Has(model.MgrRightCfgAutomations)), b(r.Has(model.MgrRightCfgAllocations)),
		b(r.Has(model.MgrRightCfgCorporate)), b(r.Has(model.MgrRightCfgPayments)),
		b(r.Has(model.MgrRightCfgMails)), b(r.Has(model.MgrRightCfgStreaming)),
		b(r.Has(model.MgrRightSrvJournals)), b(r.Has(model.MgrRightSrvReports)),
		b(r.Has(model.MgrRightCharts)), b(r.Has(model.MgrRightEmail)),
		b(r.Has(model.MgrRightNews)), b(r.Has(model.MgrRightExport)),
		b(r.Has(model.MgrRightTechsupport)), b(r.Has(model.MgrRightMarket)),
		b(r.Has(model.MgrRightAccountant)), b(r.Has(model.MgrRightAccRead)),
		b(r.Has(model.MgrRightAccDetailsName)),
		b(r.Has(model.MgrRightAccDetailsLocation)),
		b(r.Has(model.MgrRightAccDetailsAddress)),
		b(r.Has(model.MgrRightAccDetailsId)), b(r.Has(model.MgrRightAccDetailsEmail)),
		b(r.Has(model.MgrRightAccDetailsPhone)),
		b(r.Has(model.MgrRightAccDetailsGeneral)),
		b(r.Has(model.MgrRightAccTechnical)), b(r.Has(model.MgrRightAccTechModify)),
		b(r.Has(model.MgrRightAccManager)), b(r.Has(model.MgrRightAccDelete)),
		b(r.Has(model.MgrRightAccOnline)), b(r.Has(model.MgrRightConfirmActions)),
		b(r.Has(model.MgrRightNotifications)), b(r.Has(model.MgrRightTradesRead)),
		b(r.Has(model.MgrRightTradesManager)), b(r.Has(model.MgrRightTradesDelete)),
		b(r.Has(model.MgrRightTradesDealer)), b(r.Has(model.MgrRightTradesSupervisor)),
		b(r.Has(model.MgrRightQuotesRaw)), b(r.Has(model.MgrRightQuotes)),
		b(r.Has(model.MgrRightSymbolDetails)), b(r.Has(model.MgrRightRiskManager)),
		b(r.Has(model.MgrRightGroupMargin)), b(r.Has(model.MgrRightGroupCommission)),
		b(r.Has(model.MgrRightReports)), b(r.Has(model.MgrRightClientsAccess)),
		b(r.Has(model.MgrRightClientsCreate)), b(r.Has(model.MgrRightClientsEdit)),
		b(r.Has(model.MgrRightClientsDelete)), b(r.Has(model.MgrRightClientsKyc)),
		b(r.Has(model.MgrRightClientsDetailsName)),
		b(r.Has(model.MgrRightClientsDetailsLocation)),
		b(r.Has(model.MgrRightClientsDetailsAddress)),
		b(r.Has(model.MgrRightClientsDetailsId)),
		b(r.Has(model.MgrRightClientsDetailsEmail)),
		b(r.Has(model.MgrRightClientsDetailsPhone)),
		b(r.Has(model.MgrRightClientsDetailsGeneral)),
		b(r.Has(model.MgrRightDocumentsAccess)),
		b(r.Has(model.MgrRightDocumentsCreate)), b(r.Has(model.MgrRightDocumentsEdit)),
		b(r.Has(model.MgrRightDocumentsDelete)),
		b(r.Has(model.MgrRightDocumentsFilesAdd)),
		b(r.Has(model.MgrRightDocumentsFilesDelete)),
		b(r.Has(model.MgrRightCommentsAccess)), b(r.Has(model.MgrRightCommentsCreate)),
		b(r.Has(model.MgrRightCommentsDelete)))

	return err
}
