package v1

import (
	"errors"
	"time"

	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CrtClient is the create payload. Only the fields the admin panel collects at
// onboarding; everything else keeps its schema default.
type CrtClient struct {
	ClientType       int16  `json:"client_type" validate:"gte=0,lte=3"`
	ClientStatus     int16  `json:"client_status" validate:"gte=0"`
	KycStatus        int16  `json:"kyc_status" validate:"gte=0"`
	AssignedManager  *int64 `json:"assigned_manager"`
	Comment          string `json:"comment" validate:"max=4096"`
	PersonName       string `json:"person_name" validate:"required,max=128"`
	PersonMiddleName string `json:"person_middle_name" validate:"max=64"`
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

// UptClient patches a client. Pointer fields plus COALESCE, so an absent field
// keeps its value and an explicit value overwrites it.
type UptClient struct {
	ClientStatus    *int16  `json:"client_status"`
	KycStatus       *int16  `json:"kyc_status"`
	AssignedManager *int64  `json:"assigned_manager"`
	Comment         *string `json:"comment" validate:"omitempty,max=4096"`
	PersonName      *string `json:"person_name" validate:"omitempty,max=128"`
	ContactEmail    *string `json:"contact_email" validate:"omitempty,email,max=255"`
	ContactPhone    *string `json:"contact_phone" validate:"omitempty,max=64"`
	AddressCountry  *string `json:"address_country" validate:"omitempty,max=64"`
	AddressCity     *string `json:"address_city" validate:"omitempty,max=64"`
	AddressStreet   *string `json:"address_street" validate:"omitempty,max=1024"`
	AddressPostcode *string `json:"address_postcode" validate:"omitempty,max=32"`
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
	CompanyName     string `json:"company_name"`
	ContactEmail    string `json:"contact_email"`
	ContactPhone    string `json:"contact_phone"`
	AddressCountry  string `json:"address_country"`
	AddressCity     string `json:"address_city"`
	DateCreated     int64  `json:"date_created"`
	DateModified    int64  `json:"date_modified"`
}

// clientsSortable are the real columns of hst.clients, checked against the
// migration. A name that does not exist here is a 500 at request time.
var clientsSortable = utils.NewSortable(
	"client_id", "date_created", "date_modified", "person_name",
	"contact_email", "client_status", "kyc_status")

const clientColumns = `client_id, client_type, client_status, kyc_status,
	COALESCE(assigned_manager, 0), comment, person_name, company_name,
	contact_email, contact_phone, address_country, address_city,
	date_created, date_modified`

// CreateClient registers a KYC person or company.
//
// @Id			CreateClient
// @Tags		Clients
// @Accept		json
// @Produce		json
// @Success		201	{object}	ViewClient
// @Security	BearerAuth
// @Router		/api/v1/clients [post]
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
		    person_name, person_middle_name, person_birth_date, person_citizenship,
		    company_name, contact_email, contact_phone,
		    address_country, address_city, address_street, address_postcode,
		    date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$17)
		 RETURNING `+clientColumns,
		body.ClientType, body.ClientStatus, body.KycStatus, body.AssignedManager,
		body.Comment, body.PersonName, body.PersonMiddleName, body.PersonBirthDate,
		body.PersonCitizen, body.CompanyName, body.ContactEmail, body.ContactPhone,
		body.AddressCountry, body.AddressCity, body.AddressStreet, body.AddressPostcode,
		now).
		Scan(&view.ClientId, &view.ClientType, &view.ClientStatus, &view.KycStatus,
			&view.AssignedManager, &view.Comment, &view.PersonName, &view.CompanyName,
			&view.ContactEmail, &view.ContactPhone, &view.AddressCountry,
			&view.AddressCity, &view.DateCreated, &view.DateModified)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Journal(logger.TypeCfg, logger.CodeOK, "client created",
		"actor", snap.Login, "client_id", view.ClientId)

	return s.App.HttpResponseCreated(c, view)
}

// ListClients returns a page of clients.
//
// @Id			ListClients
// @Tags		Clients
// @Produce		json
// @Success		200	{array}	ViewClient
// @Security	BearerAuth
// @Router		/api/v1/clients [get]
func (s *HttpServer) ListClients(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, clientsSortable, "date_created")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	// sort_by is validated against an allowlist in QueryFilter; a bind
	// parameter cannot carry an ORDER BY clause
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+clientColumns+`
		   FROM hst.clients
		  WHERE ($1 = '' OR person_name ILIKE '%'||$1||'%' OR contact_email ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewClient{}
	for rows.Next() {
		var v ViewClient
		if err := rows.Scan(&v.ClientId, &v.ClientType, &v.ClientStatus, &v.KycStatus,
			&v.AssignedManager, &v.Comment, &v.PersonName, &v.CompanyName,
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
// @Id			GetClient
// @Tags		Clients
// @Produce		json
// @Success		200	{object}	ViewClient
// @Security	BearerAuth
// @Router		/api/v1/clients/{id} [get]
func (s *HttpServer) GetClient(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	view := &ViewClient{}
	err = s.DB.DB.QueryRow(c.UserContext(),
		`SELECT `+clientColumns+` FROM hst.clients WHERE client_id = $1`, id).
		Scan(&view.ClientId, &view.ClientType, &view.ClientStatus, &view.KycStatus,
			&view.AssignedManager, &view.Comment, &view.PersonName, &view.CompanyName,
			&view.ContactEmail, &view.ContactPhone, &view.AddressCountry,
			&view.AddressCity, &view.DateCreated, &view.DateModified)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, view)
}

// UpdateClient patches the fields present in the body.
//
// @Id			UpdateClient
// @Tags		Clients
// @Accept		json
// @Produce		json
// @Success		200	{object}	ViewClient
// @Security	BearerAuth
// @Router		/api/v1/clients/{id} [patch]
func (s *HttpServer) UpdateClient(c *fiber.Ctx) error {
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

	snap, _ := utils.GetClient(c)

	view := &ViewClient{}
	err = s.DB.DB.QueryRow(c.UserContext(),
		`UPDATE hst.clients SET
		    client_status    = COALESCE($2, client_status),
		    kyc_status       = COALESCE($3, kyc_status),
		    assigned_manager = COALESCE($4, assigned_manager),
		    comment          = COALESCE($5, comment),
		    person_name      = COALESCE($6, person_name),
		    contact_email    = COALESCE($7, contact_email),
		    contact_phone    = COALESCE($8, contact_phone),
		    address_country  = COALESCE($9, address_country),
		    address_city     = COALESCE($10, address_city),
		    address_street   = COALESCE($11, address_street),
		    address_postcode = COALESCE($12, address_postcode),
		    date_modified    = $13
		  WHERE client_id = $1
		 RETURNING `+clientColumns,
		id, body.ClientStatus, body.KycStatus, body.AssignedManager, body.Comment,
		body.PersonName, body.ContactEmail, body.ContactPhone, body.AddressCountry,
		body.AddressCity, body.AddressStreet, body.AddressPostcode,
		time.Now().UnixNano()).
		Scan(&view.ClientId, &view.ClientType, &view.ClientStatus, &view.KycStatus,
			&view.AssignedManager, &view.Comment, &view.PersonName, &view.CompanyName,
			&view.ContactEmail, &view.ContactPhone, &view.AddressCountry,
			&view.AddressCity, &view.DateCreated, &view.DateModified)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Journal(logger.TypeCfg, logger.CodeOK, "client updated",
		"actor", snap.Login, "client_id", view.ClientId)

	return s.App.HttpResponseOK(c, view)
}

// DeleteClient removes a client.
//
// client_id is ON DELETE SET NULL on users, so deleting a client that still has
// trading accounts would silently orphan them. That is refused unless the
// caller asks for it explicitly.
//
// @Id			DeleteClient
// @Tags		Clients
// @Produce		json
// @Security	BearerAuth
// @Router		/api/v1/clients/{id} [delete]
func (s *HttpServer) DeleteClient(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	if !c.QueryBool("force", false) {
		var users int
		if err := s.DB.DB.QueryRow(ctx,
			`SELECT count(*) FROM hst.users WHERE client_id = $1`, id).Scan(&users); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if users > 0 {
			return s.App.HttpResponseConflict(c, errs.ErrDeleteWhileNotEmpty)
		}
	}

	tag, err := s.DB.DB.Exec(ctx, `DELETE FROM hst.clients WHERE client_id = $1`, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Journal(logger.TypeCfg, logger.CodeWarn, "client deleted",
		"actor", snap.Login, "client_id", id)

	return s.App.HttpResponseNoContent(c)
}
