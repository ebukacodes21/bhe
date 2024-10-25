package db

import (
	"context"
	// "fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferFunds(t *testing.T) {
	store := NewRepository(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// run a concurrent transferFunds
	n := 5
	amount := int64(10)

	// channel to receive feedback
	errChannel := make(chan error)
	resultsChannel := make(chan TransferFundsResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferFunds(context.Background(), TransferFundsParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			// pass feedback to main routine
			errChannel <- err
			resultsChannel <- result
		}()
	}

	exist := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errChannel
		require.NoError(t, err)

		result := <-resultsChannel
		require.NotEmpty(t, result)

		// check tx
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// check if transfer was created
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// sender entry
		senderEntry := result.SenderEntry
		require.NotEmpty(t, senderEntry)
		require.Equal(t, account1.ID, senderEntry.AccountID)
		require.Equal(t, -amount, senderEntry.Amount)
		require.NotZero(t, senderEntry.ID)
		require.NotZero(t, senderEntry.CreatedAt)

		// check if sender entry was created
		_, err = store.GetEntry(context.Background(), senderEntry.ID)
		require.NoError(t, err)

		// receiver entry
		receiverEntry := result.ReceiverEntry
		require.NotEmpty(t, receiverEntry)
		require.Equal(t, account2.ID, receiverEntry.AccountID)
		require.Equal(t, amount, receiverEntry.Amount)
		require.NotZero(t, receiverEntry.ID)
		require.NotZero(t, receiverEntry.CreatedAt)

		// check if receiver entry was created
		_, err = store.GetEntry(context.Background(), receiverEntry.ID)
		require.NoError(t, err)

		// check account balance sender
		sender := result.Sender
		require.NotEmpty(t, sender)
		require.Equal(t, account1.ID, sender.ID)

		// check account balance receiver
		receiver := result.Receiver
		require.NotEmpty(t, receiver)
		require.Equal(t, account2.ID, receiver.ID)

		// fmt.Println(sender.Balance, receiver.Balance)
		diff1 := account1.Balance - sender.Balance
		diff2 := receiver.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, exist, k)
		exist[k] = true
	}

	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updateAccount2.Balance)
}

func TestTransferDeadlock(t *testing.T) {
	store := NewRepository(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// Run a concurrent transferFunds
	n := 10
	amount := int64(10)
	errChannel := make(chan error)

	for i := 0; i < n; i++ {
		senderId := account1.ID
		receiverId := account2.ID

		if i%2 == 1 {
			senderId = account2.ID
			receiverId = account1.ID
		}

		go func(senderId, receiverId int64) {
			_, err := store.TransferFunds(context.Background(), TransferFundsParams{
				FromAccountID: senderId,
				ToAccountID:   receiverId,
				Amount:        amount,
			})

			// Pass feedback to main routine
			errChannel <- err
		}(senderId, receiverId)
	}

	for i := 0; i < n; i++ {
		err := <-errChannel
		require.NoError(t, err)
	}

	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updateAccount1.Balance)
	require.Equal(t, account2.Balance, updateAccount2.Balance)
}
