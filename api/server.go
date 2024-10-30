package api

import (
	db "bhe/db/sqlc"
	"bhe/helper"
	"bhe/token"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     helper.Config
	repository db.Repository
	router     *gin.Engine
	tokenMaker token.TokenMaker
}

func NewServer(config helper.Config, r db.Repository) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create token maker%w", err)
	}

	server := &Server{
		repository: r,
		tokenMaker: tokenMaker,
		config:     config,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(s.tokenMaker))

	authRoutes.POST("/accounts", s.createAccount)
	authRoutes.GET("/accounts/:id", s.getAccount)
	authRoutes.GET("/accounts", s.getAccounts)

	authRoutes.POST("/transfers", s.createTransfer)

	s.router = router
}

func errorRes(err error) gin.H {
	log.Println(err)
	return gin.H{"error": err}
}
