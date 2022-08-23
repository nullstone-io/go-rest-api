package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type ResourceKeyParserFunc[T any] func(r *http.Request) (T, error)

func PathParameterKeyParser(attr string) ResourceKeyParserFunc[string] {
	return func(r *http.Request) (string, error) {
		return mux.Vars(r)[attr], nil
	}
}

type Resource[TKey any, T any] struct {
	DataAccess DataAccess[TKey, T]

	// KeyParser allows the resource to parse the unique key from the request
	KeyParser ResourceKeyParserFunc[TKey]

	// BeforeExec allows the resource to introspect the request and adjust the input payload
	BeforeExec func(r *http.Request, obj *T)
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
		if r.BeforeExec != nil {
			r.BeforeExec(req, &payload)
		}
		if tempPayload.UseExisting {
			return r.DataAccess.Ensure(payload)
		}
		return r.DataAccess.Create(payload)
	})
}

func (r Resource[TKey, T]) Get(w http.ResponseWriter, req *http.Request) {
	key, err := r.KeyParser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Op(w, req, func() (*T, error) {
		return r.DataAccess.Read(key)
	})
}

func (r Resource[TKey, T]) Update(w http.ResponseWriter, req *http.Request) {
	payload, ok := DecodeBody[T](w, req)
	if !ok {
		return
	}
	key, err := r.KeyParser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Op(w, req, func() (*T, error) {
		if r.BeforeExec != nil {
			r.BeforeExec(req, &payload)
		}
		if ok, err := r.DataAccess.Exists(payload); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}
		return r.DataAccess.Update(key, payload)
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

	key, err := r.KeyParser(req)
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
