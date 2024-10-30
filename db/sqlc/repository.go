package db

import (
	"context"
	"database/sql"
)

// contract to be implemented by both mock and real db
type Repository interface {
	Querier
	TransferFunds(ctx context.Context, args TransferFundsParams) (TransferFundsResult, error)
}

// use to run db queries and transactions
type SQLRepository struct {
	*Queries
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &SQLRepository{db: db, Queries: New(db)}
}

func (s *SQLRepository) execTx(ctx context.Context, fn func(tx *Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}

type TransferFundsParams struct {
	FromAccountID int64 `db:"from_account_id" json:"from_account_id"`
	ToAccountID   int64 `db:"to_account_id" json:"to_account_id"`
	Amount        int64 `db:"amount" json:"amount"`
}

type TransferFundsResult struct {
	Transfer      Transfer `json:"transfer"`
	Sender        Account  `json:"sender"`
	Receiver      Account  `json:"receiver"`
	SenderEntry   Entry    `json:"sender_entry"`
	ReceiverEntry Entry    `json:"receiver_entry"`
}

// transaction
func (s *SQLRepository) TransferFunds(ctx context.Context, args TransferFundsParams) (TransferFundsResult, error) {
	var result TransferFundsResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error
		// create a transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: args.FromAccountID,
			ToAccountID:   args.ToAccountID,
			Amount:        args.Amount,
		})

		if err != nil {
			return err
		}

		// create a sender entry
		result.SenderEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.FromAccountID,
			Amount:    -args.Amount,
		})

		if err != nil {
			return err
		}

		// create a receiver entry
		result.ReceiverEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.ToAccountID,
			Amount:    args.Amount,
		})

		if err != nil {
			return err
		}

		// avoid database locking at this point - order of the queries matter
		// if condition passes, debit sender first and credit receiver - vice versa
		if args.FromAccountID < args.ToAccountID {
			result.Sender, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     args.FromAccountID,
				Amount: -args.Amount,
			})

			if err != nil {
				return err
			}

			result.Receiver, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     args.ToAccountID,
				Amount: args.Amount,
			})

			if err != nil {
				return err
			}
		} else {
			// add amount
			result.Receiver, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     args.ToAccountID,
				Amount: args.Amount,
			})

			if err != nil {
				return err
			}

			result.Sender, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     args.FromAccountID,
				Amount: -args.Amount,
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
