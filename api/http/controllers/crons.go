package controllers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/crons"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"gorm.io/gorm"
)

func (h *Handler) ListCrons(c *gin.Context) {
	rows := h.crons.List()
	if apiutils.GetUserRole(c) != models.RoleAdmin {
		for i := range rows {
			rows[i].Payload = nil
		}
	}
	h.response.Success(c, "Crons", gin.H{"timezone": h.crons.Timezone(), "crons": rows})
}
func (h *Handler) ListDeletedCrons(c *gin.Context) {
	rows, err := h.crons.ListDeleted(c.Request.Context())
	if err != nil {
		h.response.ServerError(c, err)
		return
	}
	if apiutils.GetUserRole(c) != models.RoleAdmin {
		for i := range rows {
			rows[i].Payload = nil
		}
	}
	h.response.Success(c, "Deleted crons", rows)
}
func (h *Handler) PermanentlyDeleteCron(c *gin.Context) {
	id, err := cronID(c)
	if err != nil {
		h.response.BadRequest(c, "Invalid cron ID", err)
		return
	}
	err = h.crons.PermanentlyDelete(c.Request.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.response.NotFound(c, "Deleted cron not found", err)
		return
	}
	if err != nil {
		h.response.ServerError(c, err)
		return
	}
	h.response.Success(c, "Cron and its run records permanently deleted", nil)
}
func (h *Handler) GetCron(c *gin.Context) {
	id, err := cronID(c)
	if err != nil {
		h.response.BadRequest(c, "Invalid cron ID", err)
		return
	}
	cron, err := h.crons.Get(c.Request.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.response.NotFound(c, "Cron not found", err)
		return
	}
	if err != nil {
		h.response.ServerError(c, err)
		return
	}
	if apiutils.GetUserRole(c) != models.RoleAdmin {
		cron.Payload = nil
	}
	h.response.Success(c, "Cron", cron)
}
func cronID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("cron"), 10, 32)
	if err == nil && id == 0 {
		err = strconv.ErrSyntax
	}
	return uint(id), err
}
func (h *Handler) SaveCron(c *gin.Context) {
	var id uint
	var err error
	if c.Param("cron") != "" {
		id, err = cronID(c)
		if err != nil {
			h.response.BadRequest(c, "Invalid cron ID", err)
			return
		}
	}
	var in crons.Input
	if err = c.ShouldBindJSON(&in); err != nil {
		h.response.BadRequest(c, "Invalid cron configuration", err)
		return
	}
	saved, err := h.crons.Save(c.Request.Context(), id, in)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.response.NotFound(c, "Cron not found", err)
		return
	}
	if err != nil {
		if errors.Is(err, crons.ErrInvalidInput) {
			h.response.BadRequest(c, err.Error(), err)
		} else {
			h.response.ServerError(c, err)
		}
		return
	}
	h.response.Success(c, "Cron saved", saved)
}
func (h *Handler) DeleteCron(c *gin.Context) {
	id, err := cronID(c)
	if err != nil {
		h.response.BadRequest(c, "Invalid cron ID", err)
		return
	}
	if err = h.crons.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.response.NotFound(c, "Cron not found", err)
		} else {
			h.response.ServerError(c, err)
		}
		return
	}
	h.response.Success(c, "Cron deleted", nil)
}
func (h *Handler) CronRuns(c *gin.Context) {
	id, err := cronID(c)
	if err != nil {
		h.response.BadRequest(c, "Invalid cron ID", err)
		return
	}
	var before uint
	if raw := c.Query("before"); raw != "" {
		v, e := strconv.ParseUint(raw, 10, 32)
		if e != nil {
			h.response.BadRequest(c, "Invalid cursor", e)
			return
		}
		before = uint(v)
	}
	rows, err := h.crons.Runs(c.Request.Context(), id, before)
	if err != nil {
		h.response.ServerError(c, err)
		return
	}
	hasMore := len(rows) > store.CronRunPageSize
	if hasMore {
		rows = rows[:store.CronRunPageSize]
	}
	h.response.Success(c, "Cron runs", gin.H{"runs": rows, "has_more": hasMore})
}
