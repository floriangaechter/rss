package api

import (
	"log"
	"net/http"
	"text/template"
)

type PageHandler struct {
	logger *log.Logger
}

func NewPageHandler(logger *log.Logger) *PageHandler {
	return &PageHandler{
		logger: logger,
	}
}

func (h *PageHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		h.logger.Printf("ERROR: HandleHome %v", err)
		return
	}
}
