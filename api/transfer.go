package api

import (
	db "bhe/db/sqlc"
	"bhe/token"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountId int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountId   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required, currency"`
}

func (s *Server) createTransfer(ctx *gin.Context) {
	var reqBody transferRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return
	}

	fromAccount, valid := s.validAccount(ctx, reqBody.FromAccountId, reqBody.Currency)
	if !valid {
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("account does not belong to authenticated user ")
		ctx.JSON(http.StatusUnauthorized, errorRes(err))
		return
	}

	_, valid = s.validAccount(ctx, reqBody.ToAccountId, reqBody.Currency)
	if !valid {
		return
	}

	args := db.TransferFundsParams{
		FromAccountID: reqBody.FromAccountId,
		ToAccountID:   reqBody.ToAccountId,
		Amount:        reqBody.Amount,
	}

	result, err := s.repository.TransferFunds(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (s *Server) validAccount(ctx *gin.Context, accountId int64, currency string) (db.Account, bool) {
	account, err := s.repository.GetAccount(ctx, accountId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorRes(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorRes(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account currency mismatch")
		ctx.JSON(http.StatusBadRequest, errorRes(err))
		return account, false
	}

	return account, true
}
