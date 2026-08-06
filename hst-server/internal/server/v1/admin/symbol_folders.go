package admin

import (
	"context"
	"fmt"
	v1 "hstserver/internal/server/v1"
	"strings"
	"time"
	"unicode"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ViewSymbolFolder is one navigator group with no instruments yet, or listed for completeness.
type ViewSymbolFolder struct {
	Path string `json:"path"`
}

// CrtSymbolFolder creates an empty group under parent (empty parent = root under Symbols).
type CrtSymbolFolder struct {
	Parent string `json:"parent" validate:"max=255"`
	Name   string `json:"name" validate:"required,max=64"`
}

// UptSymbolFolder renames a group and every symbol beneath it.
type UptSymbolFolder struct {
	From string `json:"from" validate:"required,max=255"`
	To   string `json:"to" validate:"required,max=255"`
}

// DelSymbolFolder removes a group and, when asked, every symbol under it.
type DelSymbolFolder struct {
	Path    string `json:"path" validate:"required,max=255"`
	Cascade bool   `json:"cascade"`
}

// validateFolderSegment is the MT5 rule for a single group name: letters, digits, and . _ & # only.
func validateFolderSegment(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errs.ErrRequiredParams
	}
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if strings.ContainsRune("._&#", r) {
			continue
		}
		return fmt.Errorf("folder name %q contains a disallowed character", name)
	}
	return nil
}

// validateSymbolFolderPath checks a full backslash path used in the navigator.
func validateSymbolFolderPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" || strings.HasSuffix(path, `\`) {
		return errs.ErrRequiredParams
	}
	for _, part := range strings.Split(path, `\`) {
		if part == "" {
			return fmt.Errorf("folder path %q has an empty segment", path)
		}
		if err := validateFolderSegment(part); err != nil {
			return err
		}
	}
	return nil
}

func joinSymbolFolder(parent, name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := validateFolderSegment(name); err != nil {
		return "", err
	}
	parent = strings.Trim(parent, `\ `)
	if parent == "" {
		return name, nil
	}
	if err := validateSymbolFolderPath(parent); err != nil {
		return "", err
	}
	return parent + `\` + name, nil
}

func normalizeFolderPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if err := validateSymbolFolderPath(path); err != nil {
		return "", err
	}
	return path, nil
}

// sqlPathUnderFolder matches symbols (or folder rows) stored under folder $1.
// Paths use backslash separators (e.g. Forex.1\USDHUF.1). PostgreSQL LIKE treats \ as
// the default escape, so LIKE $1 || E'\\%' matches a literal trailing % — not what we want.
const sqlPathUnderFolder = `starts_with(path, $1 || E'\\')`

// folderExists is true when the path is stored explicitly or implied by a symbol beneath it.
func folderExists(ctx context.Context, db *pgxpool.Pool, path string) (bool, error) {
	var n int
	err := db.QueryRow(ctx,
		`SELECT (
		    EXISTS (SELECT 1 FROM hst.symbol_folders WHERE path = $1)
		    OR EXISTS (SELECT 1 FROM hst.symbols WHERE `+sqlPathUnderFolder+`)
		 )::int`, path).Scan(&n)
	return n > 0, err
}

// ListSymbolFolders returns every empty navigator group persisted on the server.
//
//	@Id			ListSymbolFolders
//	@Tags		Symbols
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewSymbolFolder}
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/folders [get]
func (s *Server) ListSymbolFolders(c *fiber.Ctx) error {
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT path FROM hst.symbol_folders ORDER BY path`)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewSymbolFolder{}
	for rows.Next() {
		var v ViewSymbolFolder
		if err := rows.Scan(&v.Path); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, out)
}

// CreateSymbolFolder adds an empty group the navigator can show before any symbol is created.
//
//	@Id			CreateSymbolFolder
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtSymbolFolder	true	"parent path and segment name"
//	@Success	201		{object}	Response{data=ViewSymbolFolder}
//	@Failure	400		{object}	Response
//	@Failure	409		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/folders [post]
func (s *Server) CreateSymbolFolder(c *fiber.Ctx) error {
	var body CrtSymbolFolder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	path, err := joinSymbolFolder(body.Parent, body.Name)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	ctx := c.UserContext()
	exists, err := folderExists(ctx, s.DB.DB, path)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if exists {
		return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
	}

	now := time.Now().UnixNano()
	if _, err := s.DB.DB.Exec(ctx,
		`INSERT INTO hst.symbol_folders (path, updated_at) VALUES ($1, $2)`, path, now); err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "symbol folder created", "actor", snap.Login, "path", path)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK,
		fmt.Sprintf("%d: symbol folder %s created", snap.Login, path), ViewSymbolFolder{Path: path})

	return s.App.HttpResponseCreated(c, ViewSymbolFolder{Path: path})
}

// RenameSymbolFolder moves a group and every symbol beneath it to a new path prefix.
//
//	@Id			RenameSymbolFolder
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		body	body		UptSymbolFolder	true	"old and new folder paths"
//	@Success	200		{object}	Response{data=ViewSymbolFolder}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/folders [patch]
func (s *Server) RenameSymbolFolder(c *fiber.Ctx) error {
	var body UptSymbolFolder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	from, err := normalizeFolderPath(body.From)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	to, err := normalizeFolderPath(body.To)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if from == to {
		return s.App.HttpResponseOK(c, ViewSymbolFolder{Path: to})
	}
	if strings.HasPrefix(to+`\`, from+`\`) || strings.HasPrefix(from+`\`, to+`\`) {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("cannot rename a folder into its own subtree"))
	}

	ctx := c.UserContext()
	exists, err := folderExists(ctx, s.DB.DB, from)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !exists {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	destExists, err := folderExists(ctx, s.DB.DB, to)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if destExists {
		return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UnixNano()

	if _, err := tx.Exec(ctx,
		`UPDATE hst.symbols
		    SET path = ($2::text || substring(path FROM length($1) + 1)),
		        date_modified = $3
		  WHERE `+sqlPathUnderFolder, from, to, now); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hst.symbol_folders SET path = $2, updated_at = $3 WHERE path = $1`, from, to, now); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hst.symbol_folders
		    SET path = ($2::text || substring(path FROM length($1) + 1)),
		        updated_at = $3
		  WHERE `+sqlPathUnderFolder, from, to, now); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := tx.Commit(ctx); err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "symbol folder renamed",
		"actor", snap.Login, "from", from, "to", to)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeOK,
		fmt.Sprintf("%d: symbol folder %s renamed to %s", snap.Login, from, to), ViewSymbolFolder{Path: to})

	return s.App.HttpResponseOK(c, ViewSymbolFolder{Path: to})
}

// DeleteSymbolFolder removes an empty group, or the whole subtree when cascade is set.
//
//	@Id			DeleteSymbolFolder
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		body	body		DelSymbolFolder	true	"folder path and whether to delete symbols beneath it"
//	@Success	204		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/folders [delete]
func (s *Server) DeleteSymbolFolder(c *fiber.Ctx) error {
	var body DelSymbolFolder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	path, err := normalizeFolderPath(body.Path)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	ctx := c.UserContext()
	exists, err := folderExists(ctx, s.DB.DB, path)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if !exists {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	var symbolCount int
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT count(*) FROM hst.symbols WHERE `+sqlPathUnderFolder, path).Scan(&symbolCount); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if symbolCount > 0 && !body.Cascade {
		return s.App.HttpResponseConflict(c, errs.ErrDeleteWhileNotEmpty)
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	deleted := []v1.ViewSymbolRef{}

	if body.Cascade {
		rows, err := tx.Query(ctx,
			`DELETE FROM hst.symbols WHERE `+sqlPathUnderFolder+` RETURNING symbol_id, symbol, path`, path)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		for rows.Next() {
			var ref v1.ViewSymbolRef
			if err := rows.Scan(&ref.SymbolId, &ref.Symbol, &ref.Path); err != nil {
				rows.Close()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
			deleted = append(deleted, ref)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		rows.Close()
	}

	tag, err := tx.Exec(ctx,
		`DELETE FROM hst.symbol_folders WHERE path = $1 OR `+sqlPathUnderFolder, path)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 && len(deleted) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	for _, ref := range deleted {
		s.NotifyWS(model.SubjectSymbol, model.EventSymbolDeleted, ref)
		s.NotifySystem(model.SubjectSystemSymbolDeleted, ref)
		s.JournalEntry(c, logger.TypeCfg, logger.CodeWarn, journal.SymbolDeletedMsg(snap.Login, ref.Symbol), ref)
	}

	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "symbol folder deleted",
		"actor", snap.Login, "path", path, "symbols", len(deleted), "cascade", body.Cascade)
	s.JournalEntry(c, logger.TypeCfg, logger.CodeWarn,
		fmt.Sprintf("%d: symbol folder %s deleted", snap.Login, path), ViewSymbolFolder{Path: path})

	return s.App.HttpResponseNoContent(c)
}
