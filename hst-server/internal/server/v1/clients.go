package v1

import (
	"context"
	"errors"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtClient is the create payload.
type CrtClient struct {
	ClientType       int16  `json:"client_type" validate:"gte=0,lte=3"`
	ClientStatus     int16  `json:"client_status" validate:"gte=0"`
	KycStatus        int16  `json:"kyc_status" validate:"gte=0"`
	AssignedManager  *int64 `json:"assigned_manager"`
	Comment          string `json:"comment" validate:"max=4096"`
	PersonName       string `json:"person_name" validate:"required,max=128"`
	PersonMiddleName string `json:"person_middle_name" validate:"max=64"`
	PersonLastName   string `json:"person_last_name" validate:"max=64"`
	PersonBirthDate  int64  `json:"person_birth_date"`
	PersonCitizen    string `json:"person_citizenship" validate:"max=64"`
	CompanyName      string `json:"company_name" validate:"max=255"`
	ContactEmail     string `json:"contact_email" validate:"required,email,max=255"`
	ContactPhone     string `json:"contact_phone" validate:"max=64"`
	AddressCountry   string `json:"address_country" validate:"max=64"`
	AddressCity      string `json:"address_city" validate:"max=64"`
	AddressStreet    string `json:"address_street" validate:"max=1024"`
	AddressPostcode  string `json:"address_postcode" validate:"max=32"`
}

// UptClient patches a client.
type UptClient struct {
	ClientType                *model.ClientType             `json:"client_type"`
	ClientStatus              *model.ClientStatus           `json:"client_status"`
	KycStatus                 *model.KycStatus              `json:"kyc_status"`
	AssignedManager           *int64                        `json:"assigned_manager"`
	ComplianceApprovedBy      *int64                        `json:"compliance_approved_by"`
	ComplianceClientCategory  *string                       `json:"compliance_client_category"`
	ComplianceDateApproval    *int64                        `json:"compliance_date_approval"`
	ComplianceDateTermination *int64                        `json:"compliance_date_termination"`
	Comment                   *string                       `json:"comment"`
	LeadCampaign              *string                       `json:"lead_campaign"`
	LeadSource                *string                       `json:"lead_source"`
	Introducer                *int64                        `json:"introducer"`
	ClientOrigin              *model.ClientOrigin           `json:"client_origin"`
	ClientOriginLogin         *int64                        `json:"client_origin_login"`
	PersonTitle               *string                       `json:"person_title"`
	PersonName                *string                       `json:"person_name"`
	PersonMiddleName          *string                       `json:"person_middle_name"`
	PersonLastName            *string                       `json:"person_last_name"`
	PersonBirthDate           *int64                        `json:"person_birth_date"`
	PersonCitizenship         *string                       `json:"person_citizenship"`
	PersonGender              *model.Gender                 `json:"person_gender"`
	PersonTaxId               *string                       `json:"person_tax_id"`
	PersonDocumentType        *string                       `json:"person_document_type"`
	PersonDocumentNumber      *string                       `json:"person_document_number"`
	PersonDocumentDate        *int64                        `json:"person_document_date"`
	PersonDocumentExtra       *string                       `json:"person_document_extra"`
	PersonEmployment          *model.Employment             `json:"person_employment"`
	PersonIndustry            *model.ClientIndustry         `json:"person_industry"`
	PersonEducation           *model.EducationLevel         `json:"person_education"`
	PersonWealthSource        *model.WealthSource           `json:"person_wealth_source"`
	PersonAnnualIncome        *float64                      `json:"person_annual_income"`
	PersonNetWorth            *float64                      `json:"person_net_worth"`
	PersonAnnualDeposit       *float64                      `json:"person_annual_deposit"`
	CompanyName               *string                       `json:"company_name"`
	CompanyRegNumber          *string                       `json:"company_reg_number"`
	CompanyRegDate            *string                       `json:"company_reg_date"`
	CompanyRegAuthority       *string                       `json:"company_reg_authority"`
	CompanyVat                *string                       `json:"company_vat"`
	CompanyLei                *string                       `json:"company_lei"`
	CompanyLicenseNumber      *string                       `json:"company_license_number"`
	CompanyLicenseAuthority   *string                       `json:"company_license_authority"`
	CompanyCountry            *string                       `json:"company_country"`
	CompanyAddress            *string                       `json:"company_address"`
	CompanyWebsite            *string                       `json:"company_website"`
	ContactPreferred          *model.PreferredCommunication `json:"contact_preferred"`
	ContactLanguage           *string                       `json:"contact_language"`
	ContactEmail              *string                       `json:"contact_email"`
	ContactPhone              *string                       `json:"contact_phone"`
	ContactMessengers         *string                       `json:"contact_messengers"`
	ContactSocialNetworks     *string                       `json:"contact_social_networks"`
	ContactLastDate           *int64                        `json:"contact_last_date"`
	AddressCountry            *string                       `json:"address_country"`
	AddressPostcode           *string                       `json:"address_postcode"`
	AddressStreet             *string                       `json:"address_street"`
	AddressState              *string                       `json:"address_state"`
	AddressCity               *string                       `json:"address_city"`
	ExperienceFx              *model.TradingExperience      `json:"experience_fx"`
	ExperienceCfd             *model.TradingExperience      `json:"experience_cfd"`
	ExperienceFutures         *model.TradingExperience      `json:"experience_futures"`
	ExperienceStocks          *model.TradingExperience      `json:"experience_stocks"`
}

// ViewClient is what the panel renders.
type ViewClient struct {
	ClientId        int64  `json:"client_id"`
	ClientType      int16  `json:"client_type"`
	ClientStatus    int16  `json:"client_status"`
	KycStatus       int16  `json:"kyc_status"`
	AssignedManager int64  `json:"assigned_manager"`
	Comment         string `json:"comment"`
	PersonName      string `json:"person_name"`
	PersonLastName  string `json:"person_last_name"`
	CompanyName     string `json:"company_name"`
	ContactEmail    string `json:"contact_email"`
	ContactPhone    string `json:"contact_phone"`
	AddressCountry  string `json:"address_country"`
	AddressCity     string `json:"address_city"`
	DateCreated     int64  `json:"date_created"`
	DateModified    int64  `json:"date_modified"`
}

// clientsSortable are the real columns of hst.clients, checked against the migration.
var clientsSortable = utils.NewSortable(
	"client_id", "date_created", "date_modified", "person_name",
	"contact_email", "client_status", "kyc_status")

// clientAllColumns is every column, in model.Client field order.
const clientAllColumns = `
	client_id, client_type, client_status, kyc_status, COALESCE(assigned_manager, 0),
	COALESCE(compliance_approved_by, 0), compliance_client_category, compliance_date_approval,
	compliance_date_termination, comment, lead_campaign, lead_source, COALESCE(introducer, 0),
	client_origin, COALESCE(client_origin_login, 0), person_title, person_name,
	person_middle_name, person_last_name, person_birth_date, person_citizenship, person_gender,
	person_tax_id, person_document_type, person_document_number, person_document_date,
	person_document_extra, person_employment, person_industry, person_education,
	person_wealth_source, person_annual_income, person_net_worth, person_annual_deposit,
	company_name, company_reg_number, company_reg_date, company_reg_authority, company_vat,
	company_lei, company_license_number, company_license_authority, company_country,
	company_address, company_website, contact_preferred, contact_language, contact_email,
	contact_phone, contact_messengers, contact_social_networks, contact_last_date,
	address_country, address_postcode, address_street, address_state, address_city, experience_fx,
	experience_cfd, experience_futures, experience_stocks, date_created, date_modified`

// clientColumns is the short shape used by the list endpoint.
const clientColumns = `client_id, client_type, client_status, kyc_status,
	COALESCE(assigned_manager, 0), comment, person_name, person_last_name, company_name,
	contact_email, contact_phone, address_country, address_city,
	date_created, date_modified`

// CreateClient registers a KYC person or company.
//
//	@Id			CreateClient
//	@Tags		Clients
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtClient	true	"the client to create"
//	@Success	201		{object}	Response{data=ViewClient}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/clients [post]
func (s *HttpServer) CreateClient(c *fiber.Ctx) error {
	var body CrtClient
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	snap, _ := utils.GetClient(c)
	now := time.Now().UnixNano()

	view := &ViewClient{}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.clients
		   (client_type, client_status, kyc_status, assigned_manager, comment,
		    person_name, person_middle_name, person_last_name, person_birth_date,
		    person_citizenship, company_name, contact_email, contact_phone,
		    address_country, address_city, address_street, address_postcode,
		    date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$18)
		 RETURNING `+clientColumns,
		body.ClientType, body.ClientStatus, body.KycStatus, body.AssignedManager,
		body.Comment, body.PersonName, body.PersonMiddleName, body.PersonLastName,
		body.PersonBirthDate, body.PersonCitizen, body.CompanyName,
		body.ContactEmail, body.ContactPhone,
		body.AddressCountry, body.AddressCity, body.AddressStreet, body.AddressPostcode,
		now).
		Scan(&view.ClientId, &view.ClientType, &view.ClientStatus, &view.KycStatus,
			&view.AssignedManager, &view.Comment, &view.PersonName, &view.PersonLastName, &view.CompanyName,
			&view.ContactEmail, &view.ContactPhone, &view.AddressCountry,
			&view.AddressCity, &view.DateCreated, &view.DateModified)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeOK, "client created",
		"actor", snap.Login, "client_id", view.ClientId)

	s.NotifyClient(c.UserContext(), view.ClientId, model.EventClientCreated, view)
	s.NotifySystem(model.SubjectSystemClientCreated, view)
	s.JournalEntry(c, logger.CodeOK, journal.ClientCreatedMsg(snap.Login, view.ClientId), view)

	return s.App.HttpResponseCreated(c, view)
}

// ListClients returns a page of clients.
//
//	@Id			ListClients
//	@Tags		Clients
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches person_name or contact_email"
//	@Param		sort_by	query		string	false	"client_id, date_created, date_modified, person_name, contact_email, client_status, kyc_status"	Enums(client_id, date_created, date_modified, person_name, contact_email, client_status, kyc_status)
//	@Param		order	query		string	false	"asc or desc"																					Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]ViewClient}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/clients [get]
func (s *HttpServer) ListClients(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, clientsSortable, "date_created")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// a client has no group of its own: it is visible through the logins it owns
	access, args := utils.GroupAccessFor(snap.IsManager, snap.ManagerGroups, `u."group"`, 4)
	visible := `EXISTS (SELECT 1 FROM hst.users u
	                     WHERE u.client_id = hst.clients.client_id AND ` + access + `)`
	if access == "TRUE" {
		visible = "TRUE"
	}
	args = append([]any{q.Search, q.Limit, q.Offset}, args...)

	// sort_by is validated against an allowlist in QueryFilter.
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+clientColumns+`
		   FROM hst.clients
		  WHERE ($1 = '' OR person_name ILIKE '%'||$1||'%' OR contact_email ILIKE '%'||$1||'%')
		    AND `+visible+`
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewClient{}
	for rows.Next() {
		var v ViewClient
		if err := rows.Scan(&v.ClientId, &v.ClientType, &v.ClientStatus, &v.KycStatus,
			&v.AssignedManager, &v.Comment, &v.PersonName, &v.PersonLastName, &v.CompanyName,
			&v.ContactEmail, &v.ContactPhone, &v.AddressCountry, &v.AddressCity,
			&v.DateCreated, &v.DateModified); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// GetClient returns one client.
//
//	@Id			GetClient
//	@Tags		Clients
//	@Produce	json
//	@Param		id	path		int	true	"client id"
//	@Success	200	{object}	Response{data=model.Client}
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/clients/{id} [get]
func (s *HttpServer) GetClient(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	client, err := s.selectClient(c.UserContext(), int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, client)
}

// selectClient reads every column into the model.
func (s *HttpServer) selectClient(ctx context.Context, id int64) (*model.Client, error) {
	c := &model.Client{}

	err := s.DB.DB.QueryRow(ctx,
		`SELECT `+clientAllColumns+` FROM hst.clients WHERE client_id = $1`, id).
		Scan(
			&c.ClientId, &c.ClientType, &c.ClientStatus, &c.KycStatus, &c.AssignedManager,
			&c.ComplianceApprovedBy, &c.ComplianceClientCategory, &c.ComplianceDateApproval,
			&c.ComplianceDateTermination, &c.Comment, &c.LeadCampaign, &c.LeadSource, &c.Introducer,
			&c.ClientOrigin, &c.ClientOriginLogin, &c.PersonTitle, &c.PersonName, &c.PersonMiddleName,
			&c.PersonLastName, &c.PersonBirthDate, &c.PersonCitizenship, &c.PersonGender, &c.PersonTaxId,
			&c.PersonDocumentType, &c.PersonDocumentNumber, &c.PersonDocumentDate,
			&c.PersonDocumentExtra, &c.PersonEmployment, &c.PersonIndustry, &c.PersonEducation,
			&c.PersonWealthSource, &c.PersonAnnualIncome, &c.PersonNetWorth, &c.PersonAnnualDeposit,
			&c.CompanyName, &c.CompanyRegNumber, &c.CompanyRegDate, &c.CompanyRegAuthority,
			&c.CompanyVat, &c.CompanyLei, &c.CompanyLicenseNumber, &c.CompanyLicenseAuthority,
			&c.CompanyCountry, &c.CompanyAddress, &c.CompanyWebsite, &c.ContactPreferred,
			&c.ContactLanguage, &c.ContactEmail, &c.ContactPhone, &c.ContactMessengers,
			&c.ContactSocialNetworks, &c.ContactLastDate, &c.AddressCountry, &c.AddressPostcode,
			&c.AddressStreet, &c.AddressState, &c.AddressCity, &c.ExperienceFx, &c.ExperienceCfd,
			&c.ExperienceFutures, &c.ExperienceStocks, &c.DateCreated, &c.DateModified)

	return c, err
}

// UpdateClient patches the fields present in the body.
//
//	@Id			UpdateClient
//	@Tags		Clients
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"client id"
//	@Param		body	body		UptClient	true	"only the fields to change"
//	@Success	200		{object}	Response{data=model.Client}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/clients/{id} [patch]
func (s *HttpServer) UpdateClient(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptClient
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	tag, err := s.DB.DB.Exec(ctx,
		`UPDATE hst.clients SET
		    client_type = COALESCE($2, client_type), client_status = COALESCE($3, client_status),
		    kyc_status = COALESCE($4, kyc_status), assigned_manager = COALESCE($5, assigned_manager),
		    compliance_approved_by = COALESCE($6, compliance_approved_by),
		    compliance_client_category = COALESCE($7, compliance_client_category),
		    compliance_date_approval = COALESCE($8, compliance_date_approval),
		    compliance_date_termination = COALESCE($9, compliance_date_termination),
		    comment = COALESCE($10, comment), lead_campaign = COALESCE($11, lead_campaign),
		    lead_source = COALESCE($12, lead_source), introducer = COALESCE($13, introducer),
		    client_origin = COALESCE($14, client_origin),
		    client_origin_login = COALESCE($15, client_origin_login),
		    person_title = COALESCE($16, person_title), person_name = COALESCE($17, person_name),
		    person_middle_name = COALESCE($18, person_middle_name),
		    person_last_name = COALESCE($19, person_last_name),
		    person_birth_date = COALESCE($20, person_birth_date),
		    person_citizenship = COALESCE($21, person_citizenship),
		    person_gender = COALESCE($22, person_gender),
		    person_tax_id = COALESCE($23, person_tax_id),
		    person_document_type = COALESCE($24, person_document_type),
		    person_document_number = COALESCE($25, person_document_number),
		    person_document_date = COALESCE($26, person_document_date),
		    person_document_extra = COALESCE($27, person_document_extra),
		    person_employment = COALESCE($28, person_employment),
		    person_industry = COALESCE($29, person_industry),
		    person_education = COALESCE($30, person_education),
		    person_wealth_source = COALESCE($31, person_wealth_source),
		    person_annual_income = COALESCE($32, person_annual_income),
		    person_net_worth = COALESCE($33, person_net_worth),
		    person_annual_deposit = COALESCE($34, person_annual_deposit),
		    company_name = COALESCE($35, company_name),
		    company_reg_number = COALESCE($36, company_reg_number),
		    company_reg_date = COALESCE($37, company_reg_date),
		    company_reg_authority = COALESCE($38, company_reg_authority),
		    company_vat = COALESCE($39, company_vat), company_lei = COALESCE($40, company_lei),
		    company_license_number = COALESCE($41, company_license_number),
		    company_license_authority = COALESCE($42, company_license_authority),
		    company_country = COALESCE($43, company_country),
		    company_address = COALESCE($44, company_address),
		    company_website = COALESCE($45, company_website),
		    contact_preferred = COALESCE($46, contact_preferred),
		    contact_language = COALESCE($47, contact_language),
		    contact_email = COALESCE($48, contact_email),
		    contact_phone = COALESCE($49, contact_phone),
		    contact_messengers = COALESCE($50, contact_messengers),
		    contact_social_networks = COALESCE($51, contact_social_networks),
		    contact_last_date = COALESCE($52, contact_last_date),
		    address_country = COALESCE($53, address_country),
		    address_postcode = COALESCE($54, address_postcode),
		    address_street = COALESCE($55, address_street),
		    address_state = COALESCE($56, address_state), address_city = COALESCE($57, address_city),
		    experience_fx = COALESCE($58, experience_fx),
		    experience_cfd = COALESCE($59, experience_cfd),
		    experience_futures = COALESCE($60, experience_futures),
		    experience_stocks = COALESCE($61, experience_stocks),
		    date_modified = $62
		  WHERE client_id = $1`,
		id,
		body.ClientType, body.ClientStatus, body.KycStatus, body.AssignedManager,
		body.ComplianceApprovedBy, body.ComplianceClientCategory, body.ComplianceDateApproval,
		body.ComplianceDateTermination, body.Comment, body.LeadCampaign, body.LeadSource,
		body.Introducer, body.ClientOrigin, body.ClientOriginLogin, body.PersonTitle,
		body.PersonName, body.PersonMiddleName, body.PersonLastName, body.PersonBirthDate,
		body.PersonCitizenship, body.PersonGender, body.PersonTaxId, body.PersonDocumentType,
		body.PersonDocumentNumber, body.PersonDocumentDate, body.PersonDocumentExtra,
		body.PersonEmployment, body.PersonIndustry, body.PersonEducation, body.PersonWealthSource,
		body.PersonAnnualIncome, body.PersonNetWorth, body.PersonAnnualDeposit, body.CompanyName,
		body.CompanyRegNumber, body.CompanyRegDate, body.CompanyRegAuthority, body.CompanyVat,
		body.CompanyLei, body.CompanyLicenseNumber, body.CompanyLicenseAuthority,
		body.CompanyCountry, body.CompanyAddress, body.CompanyWebsite, body.ContactPreferred,
		body.ContactLanguage, body.ContactEmail, body.ContactPhone, body.ContactMessengers,
		body.ContactSocialNetworks, body.ContactLastDate, body.AddressCountry, body.AddressPostcode,
		body.AddressStreet, body.AddressState, body.AddressCity, body.ExperienceFx,
		body.ExperienceCfd, body.ExperienceFutures, body.ExperienceStocks,
		time.Now().UnixNano())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	client, err := s.selectClient(ctx, int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "client updated",
		"actor", snap.Login, "client_id", id)

	s.NotifyClient(c.UserContext(), int64(id), model.EventClientUpdated, client)
	s.NotifySystem(model.SubjectSystemClientUpdated, client)
	s.JournalEntry(c, logger.CodeOK, journal.ClientUpdatedMsg(snap.Login, int64(id)), client)

	return s.App.HttpResponseOK(c, client)
}

// DeleteClient removes a client.
//
//	@Id			DeleteClient
//	@Tags		Clients
//	@Produce	json
//	@Param		id		path		int		true	"client id"
//	@Param		force	query		bool	false	"detach users whose accounts are empty"
//	@Success	204		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/clients/{id} [delete]
func (s *HttpServer) DeleteClient(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	// a client owns N users, each with one account.
	var users, funded int
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT count(*),
		        count(*) FILTER (WHERE a.balance <> 0 OR a.credit <> 0 OR a.equity <> 0)
		   FROM hst.users u
		   LEFT JOIN hst.accounts a ON a.login = u.login
		  WHERE u.client_id = $1`, id).Scan(&users, &funded); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// money outranks force.
	if funded > 0 {
		return s.App.HttpResponseConflict(c, errs.ErrClientHasFundedAccounts)
	}
	if users > 0 && !c.QueryBool("force", false) {
		return s.App.HttpResponseConflict(c, errs.ErrDeleteWhileNotEmpty)
	}

	// read the groups while the logins still point at this client
	groups := s.ClientGroups(ctx, int64(id))

	tag, err := s.DB.DB.Exec(ctx, `DELETE FROM hst.clients WHERE client_id = $1`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "client deleted",
		"actor", snap.Login, "client_id", id, "users_detached", users)

	// the logins were detached above, so the groups are read before the delete; see NotifyClient
	ref := ViewClientRef{ClientId: int64(id)}
	s.NotifyClientIn(groups, model.EventClientDeleted, ref)
	s.NotifySystem(model.SubjectSystemClientDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.ClientDeletedMsg(snap.Login, int64(id)), ref)

	return s.App.HttpResponseNoContent(c)
}
