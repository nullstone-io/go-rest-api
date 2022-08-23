package rest

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Mock(t *testing.T, method string, url string, payload any) (*httptest.ResponseRecorder, *http.Request) {
	raw, err := json.Marshal(payload)
	require.NoError(t, err, "marshal payload")
	req, err := http.NewRequest(method, url, bytes.NewBuffer(raw))
	require.NoError(t, err, "create request")
	wr := httptest.NewRecorder()
	return wr, req
}
