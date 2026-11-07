package handlers

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/umohsamuel/eba/pkg/response"
	ginresponse "github.com/umohsamuel/eba/pkg/response/gin"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/db/migrations"
	domainerrors "github.com/umohsamuel/eba/services/auth/internal/domain/errors"
)

const migrationTokenHeader = "X-Migration-Token"

type MigrationConfig struct {
	Migrator *migrations.Migrator
	Token    string
}

type MigrationHandler struct {
	migrator *migrations.Migrator
	token    []byte
}

func NewMigrationHandler(cfg MigrationConfig) *MigrationHandler {
	return &MigrationHandler{migrator: cfg.Migrator, token: []byte(cfg.Token)}
}

func (h *MigrationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if h.migrator == nil || len(h.token) == 0 {
		return
	}

	migrationGroup := rg.Group("/migrations", h.requireToken)
	{
		migrationGroup.GET("/status", h.Status)
		migrationGroup.GET("/version", h.Version)
		migrationGroup.POST("/up", h.Up)
		migrationGroup.POST("/up-by-one", h.UpByOne)
		migrationGroup.POST("/up-to/:version", h.UpTo)
		migrationGroup.POST("/down", h.Down)
		migrationGroup.POST("/down-to/:version", h.DownTo)
	}
}

func (h *MigrationHandler) requireToken(c *gin.Context) {
	provided := []byte(c.GetHeader(migrationTokenHeader))
	if subtle.ConstantTimeCompare(provided, h.token) != 1 {
		ginresponse.HandleError(c, domainerrors.ErrTokenInvalid).Send()
		c.Abort()
		return
	}

	c.Next()
}

func migrationContext(c *gin.Context) context.Context {
	return context.WithoutCancel(c.Request.Context())
}

func versionParam(c *gin.Context) (int64, error) {
	version, err := strconv.ParseInt(c.Param("version"), 10, 64)
	if err != nil || version < 0 {
		return 0, domainerrors.ErrBadRequest.Wrap(err)
	}

	return version, nil
}

type versionResponse struct {
	Version int64 `json:"version"`
}

// Status godoc
//
//	@Summary		Migration status
//	@Description	Lists every migration and whether it is applied or pending.
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Success		200					{object}	response.Envelope[[]migrations.Status]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/status [get]
func (h *MigrationHandler) Status(c *gin.Context) {
	statuses, err := h.migrator.Status(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, statuses, "success").Send()
}

// Version godoc
//
//	@Summary		Current schema version
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Success		200					{object}	response.Envelope[versionResponse]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/version [get]
func (h *MigrationHandler) Version(c *gin.Context) {
	version, err := h.migrator.Version(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, versionResponse{Version: version}, "success").Send()
}

// Up godoc
//
//	@Summary		Apply all pending migrations
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Success		200					{object}	response.Envelope[[]migrations.Result]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/up [post]
func (h *MigrationHandler) Up(c *gin.Context) {
	results, err := h.migrator.Up(migrationContext(c))
	h.sendResults(c, results, err)
}

// UpByOne godoc
//
//	@Summary		Apply the next pending migration
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Success		200					{object}	response.Envelope[[]migrations.Result]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		409					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/up-by-one [post]
func (h *MigrationHandler) UpByOne(c *gin.Context) {
	results, err := h.migrator.UpByOne(migrationContext(c))
	h.sendResults(c, results, err)
}

// UpTo godoc
//
//	@Summary		Apply pending migrations up to a version
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Param			version				path		int		true	"Target version"
//	@Success		200					{object}	response.Envelope[[]migrations.Result]
//	@Failure		400					{object}	response.Envelope[any]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		404					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/up-to/{version} [post]
func (h *MigrationHandler) UpTo(c *gin.Context) {
	version, err := versionParam(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	results, err := h.migrator.UpTo(migrationContext(c), version)
	h.sendResults(c, results, err)
}

// Down godoc
//
//	@Summary		Roll back the latest applied migration
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Success		200					{object}	response.Envelope[[]migrations.Result]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		409					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/down [post]
func (h *MigrationHandler) Down(c *gin.Context) {
	results, err := h.migrator.Down(migrationContext(c))
	h.sendResults(c, results, err)
}

// DownTo godoc
//
//	@Summary		Roll back migrations down to a version
//	@Description	Rolls back every applied migration newer than the given version. Version 0 rolls back everything.
//	@Tags			migrations
//	@Produce		json
//	@Param			X-Migration-Token	header		string	true	"Migration token"
//	@Param			version				path		int		true	"Target version"
//	@Success		200					{object}	response.Envelope[[]migrations.Result]
//	@Failure		400					{object}	response.Envelope[any]
//	@Failure		401					{object}	response.Envelope[any]
//	@Failure		500					{object}	response.Envelope[any]
//	@Router			/migrations/down-to/{version} [post]
func (h *MigrationHandler) DownTo(c *gin.Context) {
	version, err := versionParam(c)
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	results, err := h.migrator.DownTo(migrationContext(c), version)
	h.sendResults(c, results, err)
}

func (h *MigrationHandler) sendResults(c *gin.Context, results []migrations.Result, err error) {
	if err != nil {
		ginresponse.HandleError(c, err).Send()
		return
	}

	ginresponse.HandleSuccess(c, http.StatusOK, results, "success").Send()
}
