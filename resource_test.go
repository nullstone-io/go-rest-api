package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResource(t *testing.T) {
	tests := []struct {
		name       string
		before     func(t *testing.T, users *Users)
		method     string
		path       string
		payload    any
		wantCode   int
		wantResult any
		after      func(t *testing.T, users *Users)
	}{
		{
			name:       "create/should create successfully",
			method:     http.MethodPost,
			path:       "/users",
			payload:    User{Name: "brad"},
			wantCode:   http.StatusOK,
			wantResult: User{Name: "brad"},
		},
		{
			name: "create/should result in 500 internal server error because user already exists",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:     http.MethodPost,
			path:       "/users",
			payload:    User{Name: "brad"},
			wantCode:   http.StatusInternalServerError,
			wantResult: "exists",
		},
		{
			name: "create/should use existing if user already exists",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:     http.MethodPost,
			path:       "/users",
			payload:    map[string]any{"name": "brad", "useExisting": true},
			wantCode:   http.StatusOK,
			wantResult: User{Name: "brad"},
		},
		{
			name:     "get/should return 404 if user doesn't exist",
			method:   http.MethodGet,
			path:     "/users/brad",
			wantCode: http.StatusNotFound,
		},
		{
			name: "get/should find user successfully",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:     http.MethodGet,
			path:       "/users/brad",
			wantCode:   http.StatusOK,
			wantResult: User{Name: "brad"},
		},
		{
			name:     "update/should return 404 if user doesn't exist",
			method:   http.MethodPut,
			path:     "/users/brad",
			payload:  User{Name: "brad"},
			wantCode: http.StatusNotFound,
		},
		{
			name: "update/should update user successfully",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:     http.MethodPut,
			path:       "/users/brad",
			payload:    User{Name: "brad"},
			wantCode:   http.StatusOK,
			wantResult: User{Name: "brad"},
		},
		{
			name: "delete/should delete user successfully",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:   http.MethodDelete,
			path:     "/users/brad",
			wantCode: http.StatusNoContent,
			after: func(t *testing.T, users *Users) {
				ok, _ := users.Exists(User{Name: "brad"})
				assert.False(t, ok, "user no longer exists")
			},
		},
		{
			name: "delete/should skip user delete",
			before: func(t *testing.T, users *Users) {
				users.Create(User{Name: "brad"})
			},
			method:   http.MethodDelete,
			path:     "/users/brad",
			payload:  map[string]any{"skipDestroy": true},
			wantCode: http.StatusNoContent,
			after: func(t *testing.T, users *Users) {
				ok, _ := users.Exists(User{Name: "brad"})
				assert.True(t, ok, "user still exists")
			},
		},
		{
			name:     "delete/cannot delete missing user",
			method:   http.MethodDelete,
			path:     "/users/brad",
			wantCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := &Users{}
			if test.before != nil {
				test.before(t, users)
			}
			resource := Resource[string, User]{DataAccess: users, IdPathParameter: "name"}
			wr, req := Mock(t, test.method, test.path, test.payload)

			router := mux.NewRouter()
			router.Methods(http.MethodPost).Path("/users").HandlerFunc(resource.Create)
			router.Methods(http.MethodGet).Path("/users/{name}").HandlerFunc(resource.Get)
			router.Methods(http.MethodPut).Path("/users/{name}").HandlerFunc(resource.Update)
			router.Methods(http.MethodDelete).Path("/users/{name}").HandlerFunc(resource.Delete)
			router.ServeHTTP(wr, req)

			assert.Equal(t, test.wantCode, wr.Code)
			if test.wantResult == nil {
			} else if _, ok := test.wantResult.(string); ok {
				assert.Contains(t, wr.Body.String(), test.wantResult)
			} else {
				var got User
				json.Unmarshal(wr.Body.Bytes(), &got)
				assert.Equal(t, test.wantResult, got)
			}
			if test.after != nil {
				test.after(t, users)
			}
		})
	}
}
