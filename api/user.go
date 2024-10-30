package api

import (
	db "bhe/db/sqlc"
	"bhe/helper"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pg "github.com/lib/pq"
)

type createUserParams struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Fullname string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username       string    `db:"username" json:"username"`
	FullName       string    `db:"full_name" json:"full_name"`
	Email          string    `db:"email" json:"email"`
	PasswordChange time.Time `db:"password_change" json:"password_change"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:       user.Username,
		FullName:       user.FullName,
		Email:          user.Email,
		PasswordChange: user.PasswordChange,
		CreatedAt:      user.CreatedAt,
	}
}

// create user
func (s *Server) createUser(ctx *gin.Context) {
	var reqBody createUserParams
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	hash, err := helper.HashPassword(reqBody.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	args := db.CreateUserParams{
		Username: reqBody.Username,
		Password: hash,
		FullName: reqBody.Fullname,
		Email:    reqBody.Email,
	}

	user, err := s.repository.CreateUser(ctx, args)
	if err != nil {
		if pgErr, ok := err.(*pg.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorRes(pgErr))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	res := newUserResponse(user)
	ctx.JSON(http.StatusCreated, res)
}

type loginUserRequest struct {
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type loginUserResponse struct {
	AccessToken string       `db:"access_token" json:"access_token"`
	User        userResponse `db:"user" json:"user"`
}

func (s *Server) loginUser(ctx *gin.Context) {
	var reqBody loginUserRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	user, err := s.repository.GetUser(ctx, reqBody.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRes(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	err = helper.CheckPassword(reqBody.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorRes(err))
		return
	}

	accessToken, err := s.tokenMaker.CreateToken(user.Username, s.config.TokenAccess)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	resp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, resp)
}
