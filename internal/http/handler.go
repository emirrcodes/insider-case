package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/emirrcodes/insider-case/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	league *service.LeagueService
}

func NewHandler(league *service.LeagueService) *Handler {
	return &Handler{league: league}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.health)

	api := router.Group("/api")
	{
		api.POST("/leagues/reset", h.resetLeague)
		api.GET("/leagues/current", h.currentLeague)
		api.GET("/standings", h.standings)
		api.GET("/matches", h.matches)
		api.GET("/matches/weeks/:week", h.matchesByWeek)
		api.PUT("/matches/:id/result", h.editMatchResult)
		api.POST("/simulation/play-next-week", h.playNextWeek)
		api.POST("/simulation/play-all", h.playAll)
		api.GET("/predictions", h.predictions)
	}
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) resetLeague(c *gin.Context) {
	snapshot, err := h.league.Reset(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *Handler) currentLeague(c *gin.Context) {
	snapshot, err := h.league.Snapshot(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (h *Handler) standings(c *gin.Context) {
	standings, err := h.league.Standings(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"standings": standings})
}

func (h *Handler) matches(c *gin.Context) {
	matches, err := h.league.Matches(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func (h *Handler) matchesByWeek(c *gin.Context) {
	week, err := strconv.Atoi(c.Param("week"))
	if err != nil || week <= 0 {
		writeError(c, http.StatusBadRequest, errors.New("week must be a positive integer"))
		return
	}

	matches, err := h.league.MatchesByWeek(c.Request.Context(), week)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"week": week, "matches": matches})
}

func (h *Handler) playNextWeek(c *gin.Context) {
	result, err := h.league.PlayNextWeek(c.Request.Context())
	if errors.Is(err, service.ErrLeagueFinished) {
		writeError(c, http.StatusConflict, err)
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}

	snapshot, err := h.league.Snapshot(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"played_week": result, "league": snapshot})
}

func (h *Handler) playAll(c *gin.Context) {
	results, err := h.league.PlayAll(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}

	snapshot, err := h.league.Snapshot(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"weeks": results, "league": snapshot})
}

func (h *Handler) predictions(c *gin.Context) {
	predictions, err := h.league.Predictions(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":     "Predictions are available after the 4th completed week.",
		"predictions": predictions,
	})
}

func (h *Handler) editMatchResult(c *gin.Context) {
	matchID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || matchID <= 0 {
		writeError(c, http.StatusBadRequest, errors.New("match id must be a positive integer"))
		return
	}

	var request struct {
		HomeScore int `json:"home_score"`
		AwayScore int `json:"away_score"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, errors.New("request body must contain home_score and away_score"))
		return
	}

	snapshot, err := h.league.EditMatchResult(c.Request.Context(), matchID, request.HomeScore, request.AwayScore)
	if errors.Is(err, service.ErrInvalidScore) {
		writeError(c, http.StatusBadRequest, err)
		return
	}
	if errors.Is(err, service.ErrMatchNotFound) {
		writeError(c, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func writeError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}
