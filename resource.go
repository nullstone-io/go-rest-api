package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ResourceEntity interface {
	SetId(val string)
	Create(db *sql.DB) error
	Ensure(db *sql.DB) error
	Read(db *sql.DB) error
	Exists(db *sql.DB) (bool, error)
	Update(db *sql.DB) error
	Drop(db *sql.DB) error
}

type DbBroker interface {
	// Open creates a database connection against databaseName
	// If databaseName is empty, uses Connection URL from database broker
	Open(databaseName string) (*sql.DB, error)
}

type Resource[T ResourceEntity] struct {
	Db DbBroker
	// The path parameter (i.e. {id}) contained in the path identifying the resource
	IdPathParameter string
}

func (r Resource[T]) Create(w http.ResponseWriter, req *http.Request) {
	payload, ok := DecodeBody[T](w, req)
	if !ok {
		return
	}
	tempPayload, ok := DecodeBody[struct {
		UseExisting bool `json:"useExisting"`
	}](w, req)
	if !ok {
		return
	}

	db, err := r.Db.Open("")
	if err != nil {
		http.Error(w, fmt.Sprintf("error connecting to db: %s", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	r.Op(w, req, func() (*T, error) {
		if tempPayload.UseExisting {
			return &payload, payload.Ensure(db)
		}
		return &payload, payload.Create(db)
	})
}

func (r Resource[T]) Get(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var data T
	data.SetId(vars[r.IdPathParameter])

	db, err := r.Db.Open("")
	if err != nil {
		http.Error(w, fmt.Sprintf("error connecting to db: %s", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	r.Op(w, req, func() (*T, error) {
		if err := data.Read(db); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &data, nil
	})
}

func (r Resource[T]) Update(w http.ResponseWriter, req *http.Request) {
	data, ok := DecodeBody[T](w, req)
	if !ok {
		return
	}
	vars := mux.Vars(req)
	data.SetId(vars[r.IdPathParameter])

	db, err := r.Db.Open("")
	if err != nil {
		http.Error(w, fmt.Sprintf("error connecting to db: %s", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	r.Op(w, req, func() (*T, error) {
		if ok, err := data.Exists(db); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}
		return &data, data.Update(db)
	})
}

func (r Resource[T]) Delete(w http.ResponseWriter, req *http.Request) {
	tempPayload, ok := DecodeBody[struct {
		SkipDestroy bool `json:"skipDestroy"`
	}](w, req)
	if !ok {
		return
	}

	if tempPayload.SkipDestroy {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	vars := mux.Vars(req)
	var data T
	data.SetId(vars[r.IdPathParameter])

	db, err := r.Db.Open("")
	if err != nil {
		http.Error(w, fmt.Sprintf("error connecting to db: %s", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	r.Op(w, req, func() (*T, error) {
		if err := data.Drop(db); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &data, nil
	})
}

func (r Resource[T]) Op(w http.ResponseWriter, req *http.Request, fn func() (*T, error)) {
	result, err := fn()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if result == nil {
		http.NotFound(w, req)
	} else {
		if req.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
		} else if raw, err := json.Marshal(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(raw)
		}
	}
}
