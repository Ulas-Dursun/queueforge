package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulasdursun/queueforge/internal/domain"
	"github.com/ulasdursun/queueforge/internal/queue"
	"github.com/ulasdursun/queueforge/internal/repository"
)

type JobHandler struct {
	jobRepo    *repository.JobRepository
	queue      *queue.RedisQueue
	maxRetries int
}

func NewJobHandler(
	jobRepo *repository.JobRepository,
	q *queue.RedisQueue,
	maxRetries int,
) *JobHandler {
	return &JobHandler{
		jobRepo:    jobRepo,
		queue:      q,
		maxRetries: maxRetries,
	}
}

// CreateJobRequest is the expected request body for job creation.
type CreateJobRequest struct {
	Type    string `json:"type"    binding:"required"`
	Payload string `json:"payload" binding:"required"`
}

// CreateJob godoc
// POST /api/jobs
// CreateJob godoc
// @Summary      Create a new job
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Param        request body CreateJobRequest true "Job request"
// @Success      201 {object} domain.Job
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/jobs [post]
func (h *JobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	if req.Type != "URL_FETCH" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid job type",
			"message": "supported types: URL_FETCH",
		})
		return
	}

	now := time.Now()
	job := &domain.Job{
		ID:         uuid.New(),
		Type:       req.Type,
		Payload:    req.Payload,
		Status:     domain.StatusQueued,
		Attempts:   0,
		MaxRetries: h.maxRetries,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.jobRepo.Create(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create job",
			"message": err.Error(),
		})
		return
	}

	if err := h.queue.Push(c.Request.Context(), job.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue job",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// GetJob godoc
// GET /api/jobs/:id
// @Summary      Get job by ID
// @Tags         jobs
// @Produce      json
// @Param        id path string true "Job ID"
// @Success      200 {object} domain.Job
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /api/jobs/{id} [get]
func (h *JobHandler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid job ID",
			"message": "must be a valid UUID",
		})
		return
	}

	job, err := h.jobRepo.GetByID(id)
	if err != nil {
		if err == domain.ErrJobNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": "job not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get job",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs godoc
// GET /api/jobs?status=QUEUED
// @Summary      List all jobs
// @Tags         jobs
// @Produce      json
// @Param        status query string false "Filter by status"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} map[string]string
// @Router       /api/jobs [get]
func (h *JobHandler) ListJobs(c *gin.Context) {
	status := c.Query("status")

	jobs, err := h.jobRepo.List(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list jobs",
			"message": err.Error(),
		})
		return
	}

	if jobs == nil {
		jobs = []*domain.Job{}
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(jobs),
		"jobs":  jobs,
	})
}
