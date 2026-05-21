package handlers

import (
	"hitalent-test/internal/services"
	"net/http"
)

// HTTP error messages (client-facing)
const (
	ErrInvalidRequestBody      = "invalid request body"
	ErrInvalidDepartmentID     = "invalid department id"
	ErrInvalidDepth            = "depth must be a positive integer"
	ErrDepthExceeded           = "depth cannot exceed 5"
	ErrInvalidIncludeEmployees = "include_employees must be true or false"
	ErrInvalidMode             = "mode must be 'cascade' or 'reassign'"
	ErrReassignToRequired      = "reassign_to_department_id is required when mode=reassign"
	ErrInvalidReassignTo       = "invalid reassign_to_department_id"
)

// statusCodeFromError maps service error strings to HTTP status codes and messages.
func statusCodeFromError(errMsg string) (status int, message string) {
	switch errMsg {
	case services.ErrDepartmentNotFound,
		services.ErrParentNotFound,
		services.ErrInvalidDepartment:
		return http.StatusNotFound, errMsg

	case services.ErrEmptyName,
		services.ErrNameTooLong,
		services.ErrEmptyFullName,
		services.ErrFullNameTooLong,
		services.ErrEmptyPosition,
		services.ErrPositionTooLong:
		return http.StatusBadRequest, errMsg

	case services.ErrDuplicateName,
		services.ErrCyclicMove,
		services.ErrReassignToSelf,
		services.ErrReassignToDescendant:
		return http.StatusConflict, errMsg

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// respondWithServiceError sends an appropriate HTTP response for a service error.
func respondWithServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, msg := statusCodeFromError(err.Error())
	respondWithError(w, status, msg)
}
