package rest

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeBody(t *testing.T) {
	t.Run("decodes successfully", func(t *testing.T) {
		want := map[string]string{"key1": "value1", "key2": "value2"}
		wr, req := Mock(t, http.MethodGet, "http://server/path", want)

		got, ok := DecodeBody[map[string]string](wr, req)
		assert.True(t, ok)
		assert.Equal(t, want, got)
	})

	t.Run("fails to decode bad json", func(t *testing.T) {
		payload := "here is some bad json"
		var want map[string]string

		req, err := http.NewRequest(http.MethodGet, "http://server/path", bytes.NewBufferString(payload))
		require.NoError(t, err)
		wr := httptest.NewRecorder()

		got, ok := DecodeBody[map[string]string](wr, req)
		assert.False(t, ok)
		assert.Equal(t, want, got)
		assert.Equal(t, http.StatusBadRequest, wr.Code)
	})
}
