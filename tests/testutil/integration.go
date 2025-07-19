package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

var Server *httptest.Server

type IntegrationCase[REQ, RES any] struct {
	Name         string
	Req          *REQ
	Token        string
	TokenFactory func() string
	StatusCode   int
	ExpectedErr  *httpio.ErrorResponse
	Expected     *RES
	Additional   any
	OnSuccess    func()
}

func WithBaseUrl(url string) string {
	return fmt.Sprintf("%s%s", Server.URL, url)
}

func (tc *IntegrationCase[REQ, RES]) Run(t *testing.T, method, url string, checker func(expected, actual *RES)) {
	t.Helper()
	var jsonReq []byte

	if tc.Req != nil {
		var err error
		jsonReq, err = json.Marshal(tc.Req)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, WithBaseUrl(url), bytes.NewReader(jsonReq))
	require.NoError(t, err)

	token := tc.Token
	if tc.TokenFactory != nil {
		token = tc.TokenFactory()
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close() // nolint: errcheck

	require.Equal(t, tc.StatusCode, res.StatusCode)

	if tc.Expected != nil {
		n := new(RES)
		require.NoError(t, json.NewDecoder(res.Body).Decode(n))
		checker(tc.Expected, n)
		if tc.OnSuccess != nil {
			tc.OnSuccess()
		}
	} else {
		err := &httpio.ErrorResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(err))
		require.Equal(t, tc.ExpectedErr.Error.Code, err.Error.Code)
		require.NotEmpty(t, err.Error.Message)
	}
}
