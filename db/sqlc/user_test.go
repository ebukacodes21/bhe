package db

import (
	"bhe/helper"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hash, err := helper.HashPassword(helper.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username: helper.RandomOwner(),
		Password: hash,
		FullName: helper.RandomOwner(),
		Email:    helper.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.FullName, user.FullName)

	require.True(t, user.PasswordChange.IsZero())
	require.NotZero(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user.Username, user2.Username)
	require.Equal(t, user.Password, user2.Password)
	require.Equal(t, user.FullName, user2.FullName)
	require.Equal(t, user.Email, user2.Email)
	require.WithinDuration(t, user.PasswordChange, user2.PasswordChange, time.Second)
	require.WithinDuration(t, user.CreatedAt, user2.CreatedAt, time.Second)
}
