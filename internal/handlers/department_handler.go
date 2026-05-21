package handlers

import (
	"encoding/json"
	"hitalent-test/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type UpdateDepartmentRequest struct {
	Name     *string `json:"name,omitempty"`
	ParentID *uint   `json:"parent_id,omitempty"`
}

type DepartmentHandler struct {
	service services.DepartmentService
}

func NewDepartmentHandler(service services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	dept, err := h.service.Create(r.Context(), req.Name, req.ParentID)
	if err != nil {
		respondWithServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, dept)
}

func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidDepartmentID)
		return
	}

	query := r.URL.Query()
	depth := 1
	if depthStr := query.Get("depth"); depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil || depth < 1 {
			respondWithError(w, http.StatusBadRequest, ErrInvalidDepth)
			return
		}
		if depth > 5 {
			respondWithError(w, http.StatusBadRequest, ErrDepthExceeded)
			return
		}
	}

	includeEmployees := true
	if includeStr := query.Get("include_employees"); includeStr != "" {
		includeEmployees, err = strconv.ParseBool(includeStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, ErrInvalidIncludeEmployees)
			return
		}
	}

	dept, err := h.service.GetByID(r.Context(), uint(id), depth, includeEmployees)
	if err != nil {
		respondWithServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dept)
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidDepartmentID)
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidRequestBody)
		return
	}

	dept, err := h.service.Update(r.Context(), uint(id), req.Name, req.ParentID)
	if err != nil {
		respondWithServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dept)
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidDepartmentID)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade"
	}
	if mode != "cascade" && mode != "reassign" {
		respondWithError(w, http.StatusBadRequest, ErrInvalidMode)
		return
	}

	var reassignTo *uint
	if mode == "reassign" {
		reassignStr := r.URL.Query().Get("reassign_to_department_id")
		if reassignStr == "" {
			respondWithError(w, http.StatusBadRequest, ErrReassignToRequired)
			return
		}
		reassignToUint, err := strconv.ParseUint(reassignStr, 10, 32)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, ErrInvalidReassignTo)
			return
		}
		reassignTo = new(uint)
		*reassignTo = uint(reassignToUint)
	}

	err = h.service.Delete(r.Context(), uint(id), mode, reassignTo)
	if err != nil {
		respondWithServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
