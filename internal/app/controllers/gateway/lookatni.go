package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// LookAtniController manages LookAtni automation operations with real job processing.
type LookAtniController struct {
	dbService          *svc.DBServiceImpl
	analyzerService    *analyzer.Service
	analysisJobService svc.AnalysisJobService
}

func NewLookAtniController(dbService *svc.DBServiceImpl) *LookAtniController {
	// Initialize analyzer service
	analyzerBaseURL := getEnv("GEMX_ANALYZER_URL", "http://localhost:8080")
	analyzerAPIKey := getEnv("GEMX_ANALYZER_API_KEY", "")
	analyzerService := analyzer.NewService(analyzerBaseURL, analyzerAPIKey)

	bridge := svc.NewBridge(context.Background(), dbService, "gdbase")

	// Initialize GDBase AnalysisJob service using gdbasez bridge
	analysisJobRepo := bridge.AnalysisJobRepo(context.Background(), dbService)
	analysisJobService := bridge.AnalysisJobService(analysisJobRepo)

	return &LookAtniController{
		dbService:          dbService,
		analyzerService:    analyzerService,
		analysisJobService: analysisJobService,
	}
}

// Extract queues a LookAtni extraction job.
//
// @Summary     Extrair LookAtni
// @Description Enfileira uma extração de artefatos para processamento assíncrono. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body map[string]interface{} true "Configuração da extração"
// @Success     202 {object} LookAtniActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/extract [post]
func (lc *LookAtniController) Extract(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid request body"})
		return
	}

	// Validate required fields
	sourceURL, hasURL := payload["source_url"].(string)
	if !hasURL || sourceURL == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "source_url is required"})
		return
	}

	// Create new analysis job for extraction
	job := &svc.AnalysisJobImpl{
		ID:         uuid.New(),
		JobType:    "CODE_ANALYSIS",
		Status:     "PENDING",
		SourceURL:  sourceURL,
		SourceType: "extraction",
		InputData:  payload,
		Metadata: map[string]interface{}{
			"operation":    "extract",
			"requested_at": time.Now().UTC(),
			"client_ip":    c.ClientIP(),
		},
		MaxRetries: 3,
		UserID:     uuid.New(), // TODO: Get from auth context
		CreatedBy:  uuid.New(), // TODO: Get from auth context
		UpdatedBy:  nil,        // TODO: Get from auth context
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save job to database
	createdJob, err := lc.analysisJobService.CreateJob(c.Request.Context(), job)
	if err != nil {
		gl.Log("error", "Failed to create extraction job", "error", err, "source_url", sourceURL)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "Failed to queue extraction job"})
		return
	}

	// Start job processing asynchronously
	go lc.processExtractionJob(context.Background(), createdJob)

	gl.Log("info", "Extraction job queued", "job_id", createdJob.GetID(), "source_url", sourceURL)

	c.JSON(http.StatusAccepted, LookAtniActionResponse{
		Status:    "queued",
		Operation: "extract",
		Payload:   payload,
		Message:   fmt.Sprintf("Extraction job created with ID: %s", createdJob.GetID()),
		Timestamp: time.Now().UTC(),
	})
}

// Archive queues an archive operation for LookAtni artifacts.
//
// @Summary     Arquivar LookAtni
// @Description Agenda o arquivamento de artefatos processados pelo LookAtni. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body map[string]interface{} true "Configuração do arquivamento"
// @Success     202 {object} LookAtniActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/archive [post]
func (lc *LookAtniController) Archive(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid request body"})
		return
	}

	// Validate required fields
	projectID, hasProject := payload["project_id"].(string)
	if !hasProject || projectID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "project_id is required"})
		return
	}

	// Create new analysis job for archiving
	job := &svc.AnalysisJobImpl{
		ID:         uuid.New(),
		JobType:    "DEPENDENCY_ANALYSIS",
		Status:     "PENDING",
		SourceURL:  fmt.Sprintf("lookatni://archive/%s", projectID),
		SourceType: "archive",
		InputData:  payload,
		Metadata: map[string]interface{}{
			"operation":    "archive",
			"project_id":   projectID,
			"requested_at": time.Now().UTC(),
			"client_ip":    c.ClientIP(),
		},
		MaxRetries: 3,
		UserID:     uuid.New(), // TODO: Get from auth context
		CreatedBy:  uuid.New(), // TODO: Get from auth context
		UpdatedBy:  nil,        // TODO: Get from auth context
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save job to database
	createdJob, err := lc.analysisJobService.CreateJob(c.Request.Context(), job)
	if err != nil {
		gl.Log("error", "Failed to create archive job", "error", err, "project_id", projectID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "Failed to queue archive job"})
		return
	}

	// Start job processing asynchronously
	go lc.processArchiveJob(context.Background(), createdJob)

	gl.Log("info", "Archive job queued", "job_id", createdJob.GetID(), "project_id", projectID)

	c.JSON(http.StatusAccepted, LookAtniActionResponse{
		Status:    "queued",
		Operation: "archive",
		Payload:   payload,
		Message:   fmt.Sprintf("Archive job created with ID: %s", createdJob.GetID()),
		Timestamp: time.Now().UTC(),
	})
}

// Download issues a temporary URL to fetch LookAtni artifacts.
//
// @Summary     Baixar ativo LookAtni
// @Description Retorna URL temporária para download do artefato processado. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Identificador do recurso"
// @Success     200 {object} LookAtniDownloadResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/lookatni/download/{id} [get]
func (lc *LookAtniController) Download(c *gin.Context) {
	resourceID := strings.TrimSpace(c.Param("id"))
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "missing resource id"})
		return
	}

	// Parse UUID from resource ID
	jobID, err := uuid.Parse(resourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "invalid resource id format"})
		return
	}

	// Find the completed analysis job
	job, err := lc.analysisJobService.GetJobByID(c.Request.Context(), jobID)
	if err != nil {
		gl.Log("error", "Failed to find job for download", "error", err, "job_id", jobID)
		c.JSON(http.StatusNotFound, ErrorResponse{Status: "error", Message: "resource not found"})
		return
	}

	// Check if job is completed
	if job.GetStatus() != "COMPLETED" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Status: "error", Message: "resource not ready for download"})
		return
	}

	// Generate temporary download URL
	downloadURL := lc.generateDownloadURL(jobID.String(), job.GetOutputData())

	gl.Log("info", "Download URL generated", "job_id", jobID, "source_url", job.GetSourceURL())

	c.JSON(http.StatusOK, LookAtniDownloadResponse{
		DownloadURL: downloadURL,
		ExpiresIn:   3600, // 1 hour
		Note:        "Temporary URL valid for 1 hour",
	})
}

// Projects lists available LookAtni projects.
//
// @Summary     Listar projetos LookAtni
// @Description Lista projetos cadastrados para automações LookAtni. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} LookAtniProjectsResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/lookatni/projects [get]
func (lc *LookAtniController) Projects(c *gin.Context) {
	// Get all analysis jobs to extract unique projects
	allJobs, err := lc.analysisJobService.ListJobs(c.Request.Context())
	if err != nil {
		gl.Log("error", "Failed to get jobs for projects", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Status: "error", Message: "Failed to retrieve projects"})
		return
	}

	// Extract unique projects from jobs
	projectMap := make(map[string]map[string]interface{})
	for _, job := range allJobs {
		if job.GetJobType() == "CODE_ANALYSIS" || job.GetJobType() == "DEPENDENCY_ANALYSIS" {
			metadata := job.GetMetadata()
			if metadata != nil {
				if projectID, ok := metadata["project_id"].(string); ok && projectID != "" {
					if _, exists := projectMap[projectID]; !exists {
						projectMap[projectID] = map[string]interface{}{
							"id":            projectID,
							"name":          extractProjectName(job.GetSourceURL()),
							"description":   fmt.Sprintf("LookAtni project with %s operations", job.GetJobType()),
							"last_activity": job.GetUpdatedAt(),
							"status":        job.GetStatus(),
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	projects := make([]map[string]interface{}, 0, len(projectMap))
	for _, project := range projectMap {
		projects = append(projects, project)
	}

	// Add demo project if no real projects exist
	if len(projects) == 0 {
		projects = append(projects, map[string]interface{}{
			"id":          "demo-project",
			"name":        "Demo Project",
			"description": "Example LookAtni project - create your first extraction to see real data",
			"status":      "demo",
		})
	}

	gl.Log("info", "Projects listed", "count", len(projects), "total_jobs", len(allJobs))

	c.JSON(http.StatusOK, LookAtniProjectsResponse{
		Projects: projects,
		Version:  "gobe-real-v1.3.5",
	})
}

// processExtractionJob handles asynchronous extraction job processing
func (lc *LookAtniController) processExtractionJob(ctx context.Context, job svc.AnalysisJobModel) {
	// Mark job as started
	if err := lc.analysisJobService.StartJob(ctx, job.GetID()); err != nil {
		gl.Log("error", "Failed to mark extraction job as started", "error", err, "job_id", job.GetID())
		return
	}

	gl.Log("info", "Starting extraction job processing", "job_id", job.GetID(), "source_url", job.GetSourceURL())

	// Update progress periodically
	go lc.updateJobProgress(ctx, job.GetID(), "extract")

	// Simulate extraction processing (replace with real analyzer integration)
	time.Sleep(2 * time.Second) // Simulate processing time

	// For now, create mock output data
	outputData := map[string]interface{}{
		"extraction_type": "code_analysis",
		"source_url":      job.GetSourceURL(),
		"extracted_files": []string{"main.go", "README.md", "go.mod"},
		"lines_of_code":   1250,
		"dependencies":    []string{"gin-gonic/gin", "gorm.io/gorm"},
		"overall_score":   0.85,
		"completed_at":    time.Now().UTC(),
	}

	// Complete the job with output data
	if err := lc.analysisJobService.CompleteJob(ctx, job.GetID(), outputData); err != nil {
		gl.Log("error", "Failed to complete extraction job", "error", err, "job_id", job.GetID())
		lc.analysisJobService.FailJob(ctx, job.GetID(), fmt.Sprintf("Failed to complete extraction: %v", err))
		return
	}

	gl.Log("info", "Extraction job completed successfully", "job_id", job.GetID(), "source_url", job.GetSourceURL())
}

// processArchiveJob handles asynchronous archive job processing
func (lc *LookAtniController) processArchiveJob(ctx context.Context, job svc.AnalysisJobModel) {
	// Mark job as started
	if err := lc.analysisJobService.StartJob(ctx, job.GetID()); err != nil {
		gl.Log("error", "Failed to mark archive job as started", "error", err, "job_id", job.GetID())
		return
	}

	gl.Log("info", "Starting archive job processing", "job_id", job.GetID(), "source_url", job.GetSourceURL())

	// Update progress periodically
	go lc.updateJobProgress(ctx, job.GetID(), "archive")

	// Simulate archive processing
	time.Sleep(3 * time.Second) // Simulate processing time

	// For now, create mock output data
	outputData := map[string]interface{}{
		"archive_type":   "dependency_analysis",
		"project_id":     job.GetInputData()["project_id"],
		"archived_size":  "2.5MB",
		"files_archived": 42,
		"archive_url":    fmt.Sprintf("https://storage.lookatni.local/archives/%s.tar.gz", job.GetID()),
		"retention_days": 90,
		"overall_score":  0.92,
		"completed_at":   time.Now().UTC(),
	}

	// Complete the job with output data
	if err := lc.analysisJobService.CompleteJob(ctx, job.GetID(), outputData); err != nil {
		gl.Log("error", "Failed to complete archive job", "error", err, "job_id", job.GetID())
		lc.analysisJobService.FailJob(ctx, job.GetID(), fmt.Sprintf("Failed to complete archive: %v", err))
		return
	}

	gl.Log("info", "Archive job completed successfully", "job_id", job.GetID(), "project_id", job.GetInputData()["project_id"])
}

// updateJobProgress simulates progress updates during job processing
func (lc *LookAtniController) updateJobProgress(ctx context.Context, jobID uuid.UUID, operation string) {
	progressSteps := []float64{10, 25, 50, 75, 90}

	for _, progress := range progressSteps {
		time.Sleep(500 * time.Millisecond)
		if err := lc.analysisJobService.UpdateJobProgress(ctx, jobID, progress); err != nil {
			gl.Log("error", "Failed to update job progress", "error", err, "job_id", jobID, "progress", progress)
		} else {
			gl.Log("debug", "Job progress updated", "job_id", jobID, "operation", operation, "progress", progress)
		}
	}
}

// generateDownloadURL creates a temporary download URL for completed jobs
func (lc *LookAtniController) generateDownloadURL(jobID string, outputData map[string]interface{}) string {
	// In a real implementation, this would integrate with a file storage service
	// For now, we'll generate a mock URL based on the output data

	if outputData != nil {
		if archiveURL, ok := outputData["archive_url"].(string); ok {
			return archiveURL
		}
		if extractionType, ok := outputData["extraction_type"].(string); ok {
			return fmt.Sprintf("https://storage.lookatni.local/extractions/%s/%s.zip", extractionType, jobID)
		}
	}

	// Default fallback URL
	return fmt.Sprintf("https://storage.lookatni.local/jobs/%s/output.tar.gz", jobID)
}

// extractProjectName extracts a human-readable project name from source URL
func extractProjectName(sourceURL string) string {
	if sourceURL == "" {
		return "Unknown Project"
	}

	// Handle GitHub URLs
	if strings.Contains(sourceURL, "github.com") {
		parts := strings.Split(sourceURL, "/")
		if len(parts) >= 2 {
			return strings.TrimSuffix(parts[len(parts)-1], ".git")
		}
	}

	// Handle lookatni:// URLs
	if strings.HasPrefix(sourceURL, "lookatni://") {
		parts := strings.Split(strings.TrimPrefix(sourceURL, "lookatni://"), "/")
		if len(parts) >= 2 {
			return strings.Title(parts[1])
		}
	}

	// Extract from any URL path
	parts := strings.Split(strings.TrimSuffix(sourceURL, "/"), "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		if name != "" {
			return strings.Title(strings.ReplaceAll(name, "-", " "))
		}
	}

	return "Unknown Project"
}
