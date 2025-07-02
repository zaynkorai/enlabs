package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zaynkorai/enlabs/internal/transport/http"
	"github.com/zaynkorai/enlabs/pkg/config"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/zaynkorai/enlabs/docs"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
}

func NewServer(cfg *config.Config, handler *http.Handler) *Server {
	engine := gin.Default()

	engine.GET("/", getAPIBaseStatus)
	engine.GET("/health", getHealthStatus)

	engine.POST("/user/:userId/transaction", handler.ProcessTransaction)
	engine.GET("/user/:userId/balance", handler.GetUserBalance)

	engine.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &Server{
		engine: engine,
		cfg:    cfg,
	}
}

// @Summary Get API status
// @Description Get the current status of the API
// @Tags Default
// @ID get-api-status
// @Produce json
// @Success 200 {object} map[string]interface{} "Successful response with API status"
// @Router / [get]
func getAPIBaseStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to the API!",
		"status":  "running",
	})
}

// @Summary Get health status
// @Description Get the health status of the application
// @Tags Default
// @ID get-health-status
// @Produce json
// @Success 200 {object} object "Successful response with health status code and message"
// @Router /health [get]
func getHealthStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": gin.H{
			"code":    200,
			"message": "UP",
		},
	})
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.AppPort)
	log.Printf("Server starting on %s", addr)
	return s.engine.Run(addr)
}
