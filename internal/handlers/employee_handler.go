package handlers

import (
	"encoding/json"
	"hitalent-test/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type CreateEmployeeRequest struct {
	FullName string     `json:"full_name"`
	Position string     `json:"position"`
	HiredAt  *time.Time `json:"hired_at,omitempty"`
}

type EmployeeHandler struct {
	service services.EmployeeService
}

func NewEmployeeHandler(service services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	deptID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidDepartmentID)
		return
	}

	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	emp, err := h.service.Create(r.Context(), uint(deptID), req.FullName, req.Position, req.HiredAt)
	if err != nil {
		respondWithServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, emp)
}
