package v1

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"

	"github.com/jackc/pgx/v5"
)

// NewLogin is a login to open, from a manager creating one or from a public signup.
type NewLogin struct {
	ClientId         *int64
	Group            string
	Rights           int64
	Name             string
	FirstName        string
	LastName         string
	Email            string
	Phone            string
	Country          string
	City             string
	Comment          string
	PasswordMain     string
	PasswordInvestor string
	PasswordApi      string
}

// OpenLogin creates a login and the account row that holds its money, in one transaction.
func (s *HttpServer) OpenLogin(ctx context.Context, n NewLogin) (int64, int, error) {
	// the group is what the account opens with, so it is read before anything is written
	deposit, leverage, status, err := s.OpeningBalance(ctx, n.Group)
	if err != nil {
		return 0, status, err
	}

	// hash before opening the transaction.
	hashes, err := s.hashPasswords(n.PasswordMain, n.PasswordInvestor, n.PasswordApi)
	if err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	var login int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO hst.users
		   (client_id, "group", rights, name, first_name, last_name, email, phone,
		    country, city, comment, leverage,
		    password_main, password_investor, password_api,
		    registration, last_pass_change, updated_at, balance)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$16,$16,$17)
		 RETURNING login`,
		n.ClientId, n.Group, n.Rights, n.Name, n.FirstName, n.LastName,
		n.Email, n.Phone, n.Country, n.City, n.Comment, leverage,
		hashes[0], hashes[1], hashes[2], now, deposit).Scan(&login); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO hst.accounts (login, margin_leverage, balance, equity, updated_at)
		 VALUES ($1, $2, $3, $3, $4)`, login, leverage, deposit, now); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, nethttp.StatusInternalServerError, err
	}

	return login, nethttp.StatusCreated, nil
}

// hashPasswords hashes the three slots concurrently.
func (s *HttpServer) hashPasswords(passwords ...string) ([]string, error) {
	out := make([]string, len(passwords))
	errs := make([]error, len(passwords))

	var wg sync.WaitGroup
	for i, p := range passwords {
		if p == "" {
			continue
		}
		wg.Add(1)
		go func(i int, p string) {
			defer wg.Done()
			out[i], errs[i] = s.OAuth2.Hasher.HashPassword(p)
		}(i, p)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// demoSection is the group tree whose accounts open funded, as the platform names it.
const demoSection = "demo"

// preliminarySection is the tree a real signup waits in until it is approved.
const preliminarySection = "preliminary"

// defaultLeverage is what an account gets when the group names none: 1, which is no leverage at all.
const defaultLeverage int32 = 1

// OpeningBalance is what an account in this group starts with.
func (s *HttpServer) OpeningBalance(ctx context.Context, group string) (float64, int32, int, error) {
	var (
		deposit  *float64
		leverage *int32
	)

	// the group is read whichever tree it is in.
	err := s.DB.DB.QueryRow(ctx,
		`SELECT demo_deposit, demo_leverage FROM hst.groups WHERE "group" = $1`, group).
		Scan(&deposit, &leverage)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nethttp.StatusBadRequest, errs.ErrGroupNotFound
	}
	if err != nil {
		return 0, 0, nethttp.StatusInternalServerError, err
	}

	// only the demo tree opens on the house
	if !IsDemoGroup(group) {
		return 0, defaultLeverage, nethttp.StatusOK, nil
	}

	return PtrOr(deposit, 0), PtrOr(leverage, defaultLeverage), nethttp.StatusOK, nil
}

// GroupExists reports whether a group path names a group that is there.
func (s *HttpServer) GroupExists(ctx context.Context, group string) error {
	var exists bool
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM hst.groups WHERE "group" = $1)`, group).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errs.ErrGroupNotFound
	}
	return nil
}

// IsPreliminaryGroup reports whether the path opens onto the tree where accounts await approval.
func IsPreliminaryGroup(group string) bool {
	head, _, _ := strings.Cut(strings.TrimSpace(group), model.GroupSep)
	return strings.EqualFold(head, preliminarySection)
}

// IsDemoGroup reports whether the path opens onto the demo tree.
func IsDemoGroup(group string) bool {
	head, _, _ := strings.Cut(strings.TrimSpace(group), model.GroupSep)
	return strings.EqualFold(head, demoSection)
}

// ViewSymbol is the list summary of an instrument, shared by both panels.
type ViewSymbol struct {
	SymbolId     int64   `json:"symbol_id"`
	Symbol       string  `json:"symbol"`
	Path         string  `json:"path"`
	Description  string  `json:"description"`
	Digits       int32   `json:"digits"`
	TradeMode    int16   `json:"trade_mode"`
	CalcMode     int16   `json:"calc_mode"`
	ExecMode     int16   `json:"exec_mode"`
	Spread       int32   `json:"spread"`
	ContractSize float64 `json:"contract_size"`
	DateModified int64   `json:"date_modified"`
	// Swap summary — populated on admin list for copy-from-symbol without a detail fetch.
	SwapMode          int16   `json:"swap_mode"`
	SwapLong          float64 `json:"swap_long"`
	SwapShort         float64 `json:"swap_short"`
	SwapYearDay       int32   `json:"swap_year_day"`
	SwapFlags         int32   `json:"swap_flags"`
	SwapRateSunday    float64 `json:"swap_rate_sunday"`
	SwapRateMonday    float64 `json:"swap_rate_monday"`
	SwapRateTuesday   float64 `json:"swap_rate_tuesday"`
	SwapRateWednesday float64 `json:"swap_rate_wednesday"`
	SwapRateThursday  float64 `json:"swap_rate_thursday"`
	SwapRateFriday    float64 `json:"swap_rate_friday"`
	SwapRateSaturday  float64 `json:"swap_rate_saturday"`
	ColorBackground   int64   `json:"color_background"`
}

func PtrOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}
