package response

import (
	"encoding/json"
	"net/http"
)

// Error represents an API error response
type Error struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Paginated represents a paginated response
type Paginated struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// JSON sends a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// OK sends a 200 OK response
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, Error{
		Error:   "bad_request",
		Message: message,
	})
}

// NotFound sends a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, Error{
		Error:   "not_found",
		Message: message,
	})
}

// Conflict sends a 409 Conflict error
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, Error{
		Error:   "conflict",
		Message: message,
	})
}

// InternalError sends a 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, Error{
		Error:   "internal_error",
		Message: message,
	})
}

// PaginatedResponse creates a paginated response
func PaginatedResponse(w http.ResponseWriter, data interface{}, page, perPage, total int) {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	OK(w, Paginated{
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}
