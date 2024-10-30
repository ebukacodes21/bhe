package api

import (
	mockdb "bhe/db/mock"
	db "bhe/db/sqlc"
	"bhe/helper"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	cases := []struct {
		name       string
		accountId  int64
		buildStubs func(mock *mockdb.MockRepository)
		checkRes   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubs: func(repository *mockdb.MockRepository) {
				// build a stub
				repository.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkRes: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				matchBody(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			buildStubs: func(repository *mockdb.MockRepository) {
				// build a stub
				repository.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkRes: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			buildStubs: func(repository *mockdb.MockRepository) {
				// build a stub
				repository.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkRes: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidId",
			accountId: 0,
			buildStubs: func(repository *mockdb.MockRepository) {
				// build a stub
				repository.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkRes: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range cases {
		c := cases[i]

		t.Run(c.name, func(t *testing.T) {
			ctr := gomock.NewController(t)
			defer ctr.Finish()

			repository := mockdb.NewMockRepository(ctr)
			c.buildStubs(repository)

			config := helper.Config{
				TokenKey:    helper.RandomString(32),
				TokenAccess: time.Minute,
			}

			server, err := NewServer(config, repository)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", c.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			c.checkRes(t, recorder)
		})

	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       helper.RandomInt(1, 1000),
		Owner:    helper.RandomOwner(),
		Balance:  helper.RandomAmount(),
		Currency: helper.RandomCurrency(),
	}
}

func matchBody(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var rAccount db.Account
	err = json.Unmarshal(data, &rAccount)
	require.NoError(t, err)
	require.Equal(t, account, rAccount)
}
