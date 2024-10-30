package api

import (
	db "bhe/db/sqlc"
	"bhe/token"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	pg "github.com/lib/pq"
)

type createAccountParams struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (s *Server) createAccount(ctx *gin.Context) {
	var reqBody createAccountParams
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	args := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: reqBody.Currency,
	}

	account, err := s.repository.CreateAccount(ctx, args)
	if err != nil {
		if pgErr, ok := err.(*pg.Error); ok {
			switch pgErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorRes(pgErr))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	ctx.JSON(http.StatusCreated, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	account, err := s.repository.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRes(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account does not belong to this user")
		ctx.JSON(http.StatusUnauthorized, errorRes(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (s *Server) getAccounts(ctx *gin.Context) {
	var req getAccountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	args := db.GetAccountsParams{
		Owner:  authPayload.Username,
		Offset: (req.PageID - 1) * req.PageSize,
		Limit:  req.PageSize,
	}

	accounts, err := s.repository.GetAccounts(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
