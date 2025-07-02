package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zaynkorai/enlabs/internal/transport/http"
	"github.com/zaynkorai/enlabs/pkg/config"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
}

func NewServer(cfg *config.Config, handler *http.Handler) *Server {
	engine := gin.Default()

	engine.POST("/user/:userId/transaction", handler.ProcessTransaction)
	engine.GET("/user/:userId/balance", handler.GetUserBalance)

	return &Server{
		engine: engine,
		cfg:    cfg,
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.AppPort)
	log.Printf("Server starting on %s", addr)
	return s.engine.Run(addr)
}
