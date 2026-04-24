package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulasdursun/queueforge/internal/api/handler"
	"github.com/ulasdursun/queueforge/internal/api/middleware"
	"github.com/ulasdursun/queueforge/internal/queue"
	"github.com/ulasdursun/queueforge/internal/repository"
)

func NewRouter(
	jobRepo *repository.JobRepository,
	q *queue.RedisQueue,
	maxRetries int,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	r.GET("/health", handler.NewHealthHandler().Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	jobHandler := handler.NewJobHandler(jobRepo, q, maxRetries)
	api := r.Group("/api")
	{
		jobs := api.Group("/jobs")
		{
			jobs.POST("", jobHandler.CreateJob)
			jobs.GET("", jobHandler.ListJobs)
			jobs.GET("/:id", jobHandler.GetJob)
		}
	}

	return r
}
