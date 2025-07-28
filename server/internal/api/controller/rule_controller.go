package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
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

// CreateRule creates a new rule
// @Summary Create a new rule
// @Description Create a new transaction categorization rule for the authenticated user
// @Tags rules
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param rule body models.CreateRuleRequest true "Rule data"
// @Success 201 {object} models.RuleResponse "Rule created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /rule [post]
func (rc *RuleController) CreateRule(c *gin.Context) {
	var ruleReq models.CreateRuleRequest
	userId := rc.GetAuthenticatedUserId(c)
	if err := rc.BindJSON(c, &ruleReq); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	ruleReq.Rule.CreatedBy = userId
	logger.Infof("Creating new rule for user %d", ruleReq.Rule.CreatedBy)

	rule, err := rc.ruleService.CreateRule(c, ruleReq)
	if err != nil {
		logger.Errorf("Error creating rule: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule created successfully with Id %d for user %d", rule.Rule.Id, rule.Rule.CreatedBy)
	rc.SendSuccess(c, http.StatusCreated, "Rule created successfully", rule)
}

// ListRules retrieves all rules for the user
// @Summary List all rules
// @Description Get all transaction categorization rules for the authenticated user
// @Tags rules
// @Produce json
// @Security BasicAuth
// @Success 200 {array} models.RuleResponse "List of rules"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /rule [get]
func (rc *RuleController) ListRules(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Fetching all rules for user %d", userId)

	rules, err := rc.ruleService.ListRules(c, userId)
	if err != nil {
		logger.Errorf("Error fetching rules: %v", err)
		rc.HandleError(c, err)
		return
	}
	logger.Infof("Successfully fetched %d rules for user %d", len(rules), userId)
	rc.SendSuccess(c, http.StatusOK, "Rules fetched successfully", rules)
}

// GetRuleById retrieves a specific rule
// @Summary Get rule by ID
// @Description Get rule details by rule ID for the authenticated user
// @Tags rules
// @Produce json
// @Security BasicAuth
// @Param ruleId path int true "Rule ID"
// @Success 200 {object} models.RuleResponse "Rule details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Rule not found"
// @Router /rule/{ruleId} [get]
func (rc *RuleController) GetRuleById(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Fetching rule details for user %d", userId)

	ruleId, ok := rc.parseIdFromParam(c, "ruleId")
	if !ok {
		return
	}

	rule, err := rc.ruleService.GetRuleById(c, ruleId, userId)
	if err != nil {
		logger.Errorf("Error fetching rule: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule %d fetched successfully for user %d", ruleId, userId)
	rc.SendSuccess(c, http.StatusOK, "Rule fetched successfully", rule)
}

// UpdateRule updates an existing rule
// @Summary Update rule
// @Description Update rule details by rule ID for the authenticated user
// @Tags rules
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param ruleId path int true "Rule ID"
// @Param rule body models.UpdateRuleRequest true "Updated rule data"
// @Success 200 {object} models.RuleResponse "Rule updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Rule not found"
// @Router /rule/{ruleId} [patch]
func (rc *RuleController) UpdateRule(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Starting rule update for user %d", userId)

	ruleId, ok := rc.parseIdFromParam(c, "ruleId")
	if !ok {
		return
	}

	var ruleReq models.UpdateRuleRequest
	if err := rc.BindJSON(c, &ruleReq); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}

	rule, err := rc.ruleService.UpdateRule(c, ruleId, ruleReq, userId)
	if err != nil {
		logger.Errorf("Error updating rule: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule %d updated successfully for user %d", ruleId, userId)
	rc.SendSuccess(c, http.StatusOK, "Rule updated successfully", rule)
}

// UpdateRuleAction updates a rule action
// @Summary Update rule action
// @Description Update a specific action within a rule for the authenticated user
// @Tags rules
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param ruleId path int true "Rule ID"
// @Param id path int true "Action ID"
// @Param action body models.UpdateRuleActionRequest true "Updated action data"
// @Success 200 {object} models.RuleActionResponse "Rule action updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Rule or action not found"
// @Router /rule/{ruleId}/action/{id} [patch]
func (rc *RuleController) UpdateRuleAction(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Starting rule action update for user %d", userId)

	ruleId, ok := rc.parseIdFromParam(c, "ruleId")
	if !ok {
		return
	}

	id, ok := rc.parseIdFromParam(c, "id")
	if !ok {
		return
	}

	var ruleActionReq models.UpdateRuleActionRequest
	if err := rc.BindJSON(c, &ruleActionReq); err != nil {
		return
	}

	ruleAction, err := rc.ruleService.UpdateRuleAction(c, id, ruleId, ruleActionReq, userId)
	if err != nil {
		logger.Errorf("Error updating rule action: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule action %d updated successfully for user %d", id, userId)
	rc.SendSuccess(c, http.StatusOK, "Rule action updated successfully", ruleAction)
}

// UpdateRuleCondition updates a rule condition
// @Summary Update rule condition
// @Description Update a specific condition within a rule for the authenticated user
// @Tags rules
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param ruleId path int true "Rule ID"
// @Param id path int true "Condition ID"
// @Param condition body models.UpdateRuleConditionRequest true "Updated condition data"
// @Success 200 {object} models.RuleConditionResponse "Rule condition updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Rule or condition not found"
// @Router /rule/{ruleId}/condition/{id} [patch]
func (rc *RuleController) UpdateRuleCondition(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Starting rule condition update for user %d", userId)

	ruleId, ok := rc.parseIdFromParam(c, "ruleId")
	if !ok {
		return
	}

	id, ok := rc.parseIdFromParam(c, "id")
	if !ok {
		return
	}

	var ruleConditionReq models.UpdateRuleConditionRequest
	if err := rc.BindJSON(c, &ruleConditionReq); err != nil {
		return
	}

	ruleCondition, err := rc.ruleService.UpdateRuleCondition(c, id, ruleId, ruleConditionReq, userId)
	if err != nil {
		logger.Errorf("Error updating rule condition: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule condition %d updated successfully for user %d", id, userId)
	rc.SendSuccess(c, http.StatusOK, "Rule condition updated successfully", ruleCondition)
}

// DeleteRule deletes a rule
// @Summary Delete rule
// @Description Delete rule by rule ID for the authenticated user
// @Tags rules
// @Produce json
// @Security BasicAuth
// @Param ruleId path int true "Rule ID"
// @Success 204 "Rule deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Rule not found"
// @Router /rule/{ruleId} [delete]
func (rc *RuleController) DeleteRule(c *gin.Context) {
	userId := rc.GetAuthenticatedUserId(c)
	logger.Infof("Starting rule deletion for user %d", userId)

	ruleId, ok := rc.parseIdFromParam(c, "ruleId")
	if !ok {
		return
	}

	if err := rc.ruleService.DeleteRule(c, ruleId, userId); err != nil {
		logger.Errorf("Error deleting rule: %v", err)
		rc.HandleError(c, err)
		return
	}

	logger.Infof("Rule %d deleted successfully for user %d", ruleId, userId)
	rc.SendSuccess(c, http.StatusNoContent, "Rule deleted successfully", nil)
}

// parseIdFromParam retrieves an Id from a URL parameter.
// It sends an error response and returns false if parsing fails.
func (rc *RuleController) parseIdFromParam(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		rc.SendError(c, http.StatusBadRequest, "invalid "+paramName)
		return 0, false
	}
	return id, true
}
