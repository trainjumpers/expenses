package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleController struct {
	*BaseController
	ruleService service.RuleServiceInterface
}

func NewRuleController(cfg *config.Config, ruleService service.RuleServiceInterface) *RuleController {
	return &RuleController{
		BaseController: NewBaseController(cfg),
		ruleService:    ruleService,
	}
}

func (rc *RuleController) CreateRule(c *gin.Context) {
	var req models.CreateRuleRequest
	if err := rc.BindJSON(c, &req); err != nil {
		return
	}
	resp, err := rc.ruleService.CreateRule(c, &req)
	if err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusCreated, "Rule created successfully", resp)
}

func (rc *RuleController) GetAllRules(c *gin.Context) {
	rules, err := rc.ruleService.ListRules(c)
	if err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusOK, "Rules fetched successfully", rules)
}

func (rc *RuleController) GetRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rc.SendError(c, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := rc.ruleService.GetRuleByID(c, id)
	if err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusOK, "Rule fetched successfully", rule)
}

// PUT /rules/:id
func (rc *RuleController) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rc.SendError(c, http.StatusBadRequest, "invalid rule id")
		return
	}
	var req models.UpdateRuleRequest
	if err := rc.BindJSON(c, &req); err != nil {
		return
	}
	if err := rc.ruleService.UpdateRule(c, id, &req); err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusNoContent, "Rule updated successfully", nil)
}

// DELETE /rules/:id
func (rc *RuleController) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rc.SendError(c, http.StatusBadRequest, "invalid rule id")
		return
	}
	if err := rc.ruleService.DeleteRule(c, id); err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusNoContent, "Rule deleted successfully", nil)
}

// POST /rules/execute
func (rc *RuleController) ExecuteRules(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		rc.SendError(c, http.StatusBadRequest, "invalid user_id")
		return
	}
	resp, err := rc.ruleService.ExecuteRules(c, userId)
	if err != nil {
		rc.HandleError(c, err)
		return
	}
	rc.SendSuccess(c, http.StatusOK, "Rules executed successfully", resp)
}
