package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/LogicGateTech/devops-code-challenge/api"
	"github.com/LogicGateTech/devops-code-challenge/conf"
)

type Response struct {
	Code   int     `json:"code"`
	Error  string  `json:"error,omitempty"`
	Msg    *string `json:"msg,omitempty"`
	Status string  `json:"status"`
}

// http status: https://go.dev/src/net/http/status.go
func ResponseWithMsg(status int, msg string) Response {
	return Response{
		Code:   status,
		Msg:    &msg,
		Status: http.StatusText(status),
	}
}

func ResponseWithError(status int, err error) Response {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return Response{
		Code:   status,
		Error:  errMsg,
		Status: http.StatusText(status),
	}
}

type Router struct {
	api  *api.API
	conf *conf.Conf
	log  *slog.Logger
}

func New() (*Router, error) {
	var err error
	router := &Router{}

	if router.conf, err = conf.New(); err != nil {
		return nil, err
	}

	// Create logger based on configuration
	router.log = conf.NewLogger(router.conf)

	if router.api, err = api.New(); err != nil {
		return nil, err
	}
	return router, nil
}

func (r *Router) Bootstrap() {
	// Apply middleware
	logging := loggingMiddleware(r.log)
	reqID := requestID
	security := securityHeaders

	// Register routes with middleware
	http.HandleFunc("/ping", security(logging(reqID(r.pong))))
	http.HandleFunc("/health", security(logging(reqID(r.health))))
	http.HandleFunc("POST /signup", security(logging(reqID(r.signup))))
	http.HandleFunc("PUT /reset", security(logging(reqID(r.resetPassword))))
}

// Close closes all resources held by the router
func (r *Router) Close() error {
	if r.api != nil {
		return r.api.Close()
	}
	return nil
}

func (r *Router) health(w http.ResponseWriter, req *http.Request) {
	// Check database connectivity
	if r.api != nil {
		if err := r.api.HealthCheck(); err != nil {
			r.log.Error("health check failed: database ping error", "error", err)
			JSONResponse(w, ResponseWithError(http.StatusServiceUnavailable, errors.New("service unavailable")))
			return
		}
	}
	JSONResponse(w, ResponseWithMsg(http.StatusOK, "OK"))
}

func (r *Router) pong(w http.ResponseWriter, req *http.Request) {
	JSONResponse(w, ResponseWithMsg(http.StatusOK, "PONG"))
}

func (r *Router) signup(w http.ResponseWriter, req *http.Request) {
	// Validate HTTP method
	if req.Method != http.MethodPost {
		JSONResponse(w, ResponseWithError(http.StatusMethodNotAllowed, errors.New("method not allowed")))
		return
	}

	// Validate Content-Type (allow form-urlencoded and multipart)
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
		!strings.Contains(contentType, "multipart/form-data") {
		JSONResponse(w, ResponseWithError(http.StatusUnsupportedMediaType, errors.New("unsupported content type")))
		return
	}

	// Limit request body size to prevent DoS
	req.Body = http.MaxBytesReader(w, req.Body, 1024*1024) // 1MB limit

	if err := req.ParseForm(); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, errors.New("invalid form data")))
		return
	}

	var email, password string
	var foundEmail, foundPassword bool

	for key, value := range req.Form {
		if len(value) == 0 {
			continue
		}
		if strings.EqualFold(key, "username") {
			email = strings.TrimSpace(value[0])
			foundEmail = true
		}
		if strings.EqualFold(key, "password") {
			password = value[0]
			foundPassword = true
		}
	}

	if !foundEmail || !foundPassword {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, errors.New("username and password are required")))
		return
	}

	// Validate input
	if err := ValidateEmail(email); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, err))
		return
	}

	if err := ValidatePassword(password); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, err))
		return
	}

	if err := r.api.Signup(email, password); err != nil {
		// Don't expose internal error details
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			JSONResponse(w, ResponseWithError(http.StatusConflict, errors.New("user already exists")))
		} else {
			r.log.Error("signup failed", "error", err, "email", email)
			JSONResponse(w, ResponseWithError(http.StatusInternalServerError, errors.New("failed to create user")))
		}
		return
	}
	JSONResponse(w, ResponseWithMsg(http.StatusCreated, "Signup Successful"))
}

func (r *Router) resetPassword(w http.ResponseWriter, req *http.Request) {
	// Validate HTTP method
	if req.Method != http.MethodPut {
		JSONResponse(w, ResponseWithError(http.StatusMethodNotAllowed, errors.New("method not allowed")))
		return
	}

	// Validate Content-Type (allow form-urlencoded and multipart)
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
		!strings.Contains(contentType, "multipart/form-data") {
		JSONResponse(w, ResponseWithError(http.StatusUnsupportedMediaType, errors.New("unsupported content type")))
		return
	}

	// Limit request body size to prevent DoS
	req.Body = http.MaxBytesReader(w, req.Body, 1024*1024) // 1MB limit

	if err := req.ParseForm(); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, errors.New("invalid form data")))
		return
	}

	var email, password string
	var foundEmail, foundPassword bool

	for key, value := range req.Form {
		if len(value) == 0 {
			continue
		}
		if strings.EqualFold(key, "username") {
			email = strings.TrimSpace(value[0])
			foundEmail = true
		}
		if strings.EqualFold(key, "password") {
			password = value[0]
			foundPassword = true
		}
	}

	if !foundEmail || !foundPassword {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, errors.New("username and password are required")))
		return
	}

	// Validate input
	if err := ValidateEmail(email); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, err))
		return
	}

	if err := ValidatePassword(password); err != nil {
		JSONResponse(w, ResponseWithError(http.StatusBadRequest, err))
		return
	}

	if err := r.api.Reset(email, password); err != nil {
		// Don't expose whether user exists or not (security best practice)
		r.log.Error("password reset failed", "error", err, "email", email)
		// Return same error for user not found to prevent enumeration
		JSONResponse(w, ResponseWithError(http.StatusNotFound, errors.New("user not found or password update failed")))
		return
	}
	JSONResponse(w, ResponseWithMsg(http.StatusOK, "Password Updated"))
}

func JSONResponse(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(response.Code)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log encoding error but can't change response at this point
		// In production, this should be logged to monitoring system
	}
}
