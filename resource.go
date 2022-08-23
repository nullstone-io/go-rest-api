package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type Resource[TKey any, T any] struct {
	DataAccess DataAccess[TKey, T]
	// The path parameter (i.e. {id}) contained in the path identifying the resource
	IdPathParameter string
}

func (r Resource[TKey, T]) Create(w http.ResponseWriter, req *http.Request) {
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

	r.Op(w, req, func() (*T, error) {
		if tempPayload.UseExisting {
			return r.DataAccess.Ensure(payload)
		}
		return r.DataAccess.Create(payload)
	})
}

func (r Resource[TKey, T]) Get(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	key, err := r.DataAccess.ParseKey(vars[r.IdPathParameter])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Op(w, req, func() (*T, error) {
		return r.DataAccess.Read(key)
	})
}

func (r Resource[TKey, T]) Update(w http.ResponseWriter, req *http.Request) {
	data, ok := DecodeBody[T](w, req)
	if !ok {
		return
	}
	vars := mux.Vars(req)
	key, err := r.DataAccess.ParseKey(vars[r.IdPathParameter])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Op(w, req, func() (*T, error) {
		if ok, err := r.DataAccess.Exists(data); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}
		return r.DataAccess.Update(key, data)
	})
}

func (r Resource[TKey, T]) Delete(w http.ResponseWriter, req *http.Request) {
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
	key, err := r.DataAccess.ParseKey(vars[r.IdPathParameter])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Op(w, req, func() (*T, error) {
		var data T
		ok, err := r.DataAccess.Drop(key)
		if !ok {
			return nil, nil
		}
		return &data, err
	})
}

func (r Resource[TKey, T]) Op(w http.ResponseWriter, req *http.Request, fn func() (*T, error)) {
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
