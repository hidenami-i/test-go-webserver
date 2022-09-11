package main

import (
	"net/http"
	"webserver/handler"
	"webserver/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
)

func NewMux() http.Handler {
	mux := chi.NewRouter()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	v := validator.New()
	at := &handler.AddTask{
		Store:     store.Tasks,
		Validator: v,
	}
	mux.Post("/tasks", at.ServeHttp)

	lt := &handler.ListTask{Store: store.Tasks}
	mux.Get("/tasks", lt.ServeHTTP)

	return mux
}
