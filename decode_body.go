package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func DecodeBody[T any](w http.ResponseWriter, req *http.Request) (T, bool) {
	original, _ := ioutil.ReadAll(req.Body)
	defer func() {
		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(original))
	}()

	var payload T
	if err := json.Unmarshal(original, &payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %s", err), http.StatusBadRequest)
		return payload, false
	}
	return payload, true
}
