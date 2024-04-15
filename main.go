package main

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AddRequest struct {
	AdPlacementID uuid.UUID `json:"addPlacementID"`
}

type AddResponse struct {
	AddID    uuid.UUID `json:"addID"`
	BidPrice float64   `json:"bidPrice"`
}

type AddService interface {
	Add(uuid.UUID) (uuid.UUID, float64, error)
}

type loggingMiddleware struct {
	next AddService
}

func NewLoggingMiddleware(svc AddService) AddService {
	return &loggingMiddleware{
		next: svc,
	}
}

type addService struct {
}

func (addService *addService) Add(i uuid.UUID) (uuid.UUID, float64, error) {
	n := rand.Float64()
	return uuid.New(), 69.69 + n, nil
}

func NewAddService() AddService {
	return &addService{}
}

func (lm *loggingMiddleware) Add(id uuid.UUID) (uuid uuid.UUID, bidPrice float64, err error) {

	defer func(start time.Time) {
		slog.Info(
			"addrequest",
			"id", uuid,
			"bidPrice", bidPrice,
			"err", err,
			"took", time.Since(start))
	}(time.Now())
	return lm.next.Add(id)
}
func main() {

	router := http.NewServeMux()

	svc := NewLoggingMiddleware(NewAddService())
	h := NewAddRequestHandler(svc)
	router.HandleFunc("/add", makeHandler(h.handleAddRequest))
	http.ListenAndServe(":3000", router)
}

type addRequestHandler struct {
	svc AddService
}

func NewAddRequestHandler(svc AddService) *addRequestHandler {
	return &addRequestHandler{
		svc: svc,
	}
}
func handleAddRequest(w http.ResponseWriter, r *http.Request) {
	resp := AddResponse{
		BidPrice: 69.69,
		AddID:    uuid.New(),
	}
	writeJSON(w, http.StatusOK, resp)

}
func (h addRequestHandler) handleAddRequest(w http.ResponseWriter, r *http.Request) error {
	addID := uuid.New()

	id, bidPrice, err := h.svc.Add(addID)
	if err != nil {
		slog.Error("add service returned non 200 response", "err", err)
		if err := writeJSON(w, http.StatusNoContent, nil); err != nil {
		}
		return writeJSON(w, http.StatusNoContent, nil)
	}
	resp := AddResponse{
		AddID:    id,
		BidPrice: bidPrice,
	}
	return writeJSON(w, http.StatusOK, resp)
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func makeHandler(h APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Error("API error", "err", err, "path", r.URL.Path)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(&v)

}
