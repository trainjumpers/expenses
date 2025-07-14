package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/logger"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CustomParserController struct {
	*BaseController
	customParserService service.CustomParserServiceInterface
	statementService    service.StatementServiceInterface
	statementValidator  *validator.StatementValidator
}

func NewCustomParserController(
	cfg *config.Config,
	customParserService service.CustomParserServiceInterface,
	statementService service.StatementServiceInterface,
) *CustomParserController {
	return &CustomParserController{
		BaseController:      NewBaseController(cfg),
		customParserService: customParserService,
		statementService:    statementService,
		statementValidator:  validator.NewStatementValidator(),
	}
}

// PreviewStatement handles POST /statements/preview - shows file preview for column mapping
func (c *CustomParserController) PreviewStatement(ctx *gin.Context) {
	userID := c.GetAuthenticatedUserId(ctx)
	logger.Infof("Previewing statement for user %d", userID)

	// Parse multipart form
	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		c.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get form values
	accountIDStr := ctx.PostForm("account_id")
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		c.SendError(ctx, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Validate input using validator
	uploadInput, err := c.statementValidator.ValidateStatementUpload(accountIDStr, file, header)
	if err != nil {
		logger.Errorf("Statement upload validation failed: %v", err)
		c.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(uploadInput.File)
	if err != nil {
		logger.Errorf("Failed to read file content: %v", err)
		c.SendError(ctx, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Determine file type
	fileType := "csv"
	filename := strings.ToLower(uploadInput.Header.Filename)
	if strings.HasSuffix(filename, ".xls") || strings.HasSuffix(filename, ".xlsx") {
		fileType = "excel"
	}

	// Get preview
	preview, err := c.customParserService.PreviewStatement(ctx, fileBytes, fileType)
	if err != nil {
		logger.Errorf("Error previewing statement: %v", err)
		c.HandleError(ctx, err)
		return
	}

	// Create a temporary statement record for the parsing process
	input := models.CreateStatementInput{
		AccountID:        uploadInput.AccountID,
		CreatedBy:        userID,
		OriginalFilename: uploadInput.Header.Filename,
		FileType:         fileType,
		Status:           models.StatementStatusPending,
	}

	// Store the statement temporarily using ParseStatement (which creates the statement)
	statement, err := c.statementService.ParseStatement(ctx, input, fileBytes)
	if err != nil {
		logger.Errorf("Error creating temporary statement: %v", err)
		c.HandleError(ctx, err)
		return
	}

	// Store file content temporarily (in a real implementation, you might want to store this in a temporary storage)
	// For now, we'll include the statement ID in the response so the frontend can reference it

	response := map[string]interface{}{
		"statement_id": statement.ID,
		"preview":      preview,
		"account_id":   uploadInput.AccountID,
		"filename":     uploadInput.Header.Filename,
	}

	logger.Infof("Statement preview generated successfully for user %d", userID)
	c.SendSuccess(ctx, http.StatusOK, "Statement preview generated successfully", response)
}

// ParseStatement handles POST /statements/:id/parse - parses statement with user-defined column mapping
func (c *CustomParserController) ParseStatement(ctx *gin.Context) {
	userID := c.GetAuthenticatedUserId(ctx)

	// Validate statement ID
	statementID, err := c.statementValidator.ValidateStatementID(ctx.Param("id"))
	if err != nil {
		logger.Errorf("Statement ID validation failed: %v", err)
		c.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	logger.Infof("Parsing statement %d for user %d", statementID, userID)

	// Parse request body
	var request models.ParseStatementRequest
	if err := c.BindJSON(ctx, &request); err != nil {
		logger.Errorf("Failed to bind JSON: %v", err)
		return
	}

	// Validate that the statement belongs to the user
	_, err = c.statementService.GetStatementStatus(ctx, statementID, userID)
	if err != nil {
		logger.Errorf("Error fetching statement: %v", err)
		c.HandleError(ctx, err)
		return
	}

	// For this implementation, we'll need to re-read the file
	// In a production system, you might want to store the file content temporarily
	// For now, we'll return an error asking the user to re-upload
	c.SendError(ctx, http.StatusNotImplemented, "Custom parsing with stored files not yet implemented. Please use the preview and parse in a single step.")
}

// ParseStatementDirect handles POST /statements/parse-direct - preview and parse in one step
func (c *CustomParserController) ParseStatementDirect(ctx *gin.Context) {
	userID := c.GetAuthenticatedUserId(ctx)
	logger.Infof("Direct parsing statement for user %d", userID)

	// Parse multipart form
	err := ctx.Request.ParseMultipartForm(256 << 10) // 256KB max
	if err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		c.SendError(ctx, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get form values
	accountIDStr := ctx.PostForm("account_id")
	hasHeaders := ctx.PostForm("has_headers") == "true"
	
	// Parse column mapping from form
	var columnMapping models.ColumnMapping
	if err := ctx.ShouldBind(&columnMapping); err != nil {
		logger.Errorf("Failed to bind column mapping: %v", err)
		c.SendError(ctx, http.StatusBadRequest, "Invalid column mapping")
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		logger.Errorf("Failed to get file from form: %v", err)
		c.SendError(ctx, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Validate input using validator
	uploadInput, err := c.statementValidator.ValidateStatementUpload(accountIDStr, file, header)
	if err != nil {
		logger.Errorf("Statement upload validation failed: %v", err)
		c.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(uploadInput.File)
	if err != nil {
		logger.Errorf("Failed to read file content: %v", err)
		c.SendError(ctx, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Determine file type
	fileType := "csv"
	filename := strings.ToLower(uploadInput.Header.Filename)
	if strings.HasSuffix(filename, ".xls") || strings.HasSuffix(filename, ".xlsx") {
		fileType = "excel"
	}

	// Parse with mapping
	parseInput := models.ParseStatementInput{
		FileBytes:     fileBytes,
		FileType:      fileType,
		HasHeaders:    hasHeaders,
		ColumnMapping: columnMapping,
	}

	result, err := c.customParserService.ParseWithMapping(ctx, parseInput)
	if err != nil {
		logger.Errorf("Error parsing statement: %v", err)
		c.HandleError(ctx, err)
		return
	}

	// Create statement record using ParseStatement
	statementInput := models.CreateStatementInput{
		AccountID:        uploadInput.AccountID,
		CreatedBy:        userID,
		OriginalFilename: uploadInput.Header.Filename,
		FileType:         fileType,
		Status:           models.StatementStatusDone,
	}

	statement, err := c.statementService.ParseStatement(ctx, statementInput, fileBytes)
	if err != nil {
		logger.Errorf("Error creating statement: %v", err)
		c.HandleError(ctx, err)
		return
	}

	// TODO: Create transactions from parsed result
	// This would involve calling the transaction service to create transactions

	response := map[string]interface{}{
		"statement":    statement,
		"parse_result": result,
	}

	logger.Infof("Statement parsed successfully with %d transactions for user %d", len(result.Transactions), userID)
	c.SendSuccess(ctx, http.StatusCreated, "Statement parsed successfully", response)
}
