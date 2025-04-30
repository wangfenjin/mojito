package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/utils"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
	"github.com/wangfenjin/mojito/pkg/openapi"
)

var validate = validator.New()

// WithHandler creates middleware that handles both request parsing and response writing
func WithHandler[Req any, Resp any](handler func(ctx context.Context, req Req) (Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Register handler for OpenAPI if not already registered
		pattern := chi.RouteContext(ctx).RoutePattern()
		if !openapi.Registered(r.Method, pattern) {
			handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
			openapi.RegisterHandler(r.Method, pattern, handlerName, reflect.TypeOf((*Req)(nil)).Elem(), reflect.TypeOf((*Resp)(nil)).Elem())
			logger.GetLogger().Info("Registering handler", "name", handlerName, "path", pattern, "method", r.Method)
		}

		var req Req
		reqType := reflect.TypeOf(req)

		if reqType != nil && reqType.Kind() == reflect.Struct && reqType.NumField() > 0 {

			// Parse HTTP headers
			reqValue := reflect.ValueOf(&req).Elem()
			for i := 0; i < reqType.NumField(); i++ {
				field := reqType.Field(i)
				headerTag := field.Tag.Get("header")
				if headerTag != "" {
					headerValue := r.Header.Get(headerTag)
					if headerValue != "" {
						fieldValue := reqValue.Field(i)
						if fieldValue.CanSet() {
							fieldValue.SetString(headerValue)
						}
					}
				}
			}

			contentType := r.Header.Get("Content-Type")
			if contentType == "application/x-www-form-urlencoded" {
				if err := r.ParseForm(); err != nil {
					logger.GetLogger().Error("Form parse error", "error", err)
					respondWithError(w, NewBadRequestError(err.Error()))
					return
				}

				// Iterate through form values and set corresponding struct fields
				reqValue := reflect.ValueOf(&req).Elem()
				for i := 0; i < reqType.NumField(); i++ {
					field := reqType.Field(i)
					formTag := field.Tag.Get("form")
					if formTag != "" {
						// Extract the field name from the form tag (remove binding info)
						fieldName := formTag
						if idx := fieldName; idx != "" {
							formValue := r.FormValue(fieldName)
							if formValue != "" {
								fieldValue := reqValue.Field(i)
								if fieldValue.CanSet() {
									fieldValue.SetString(formValue)
								}
							}
						}
					}
				}
			} else if r.Method != http.MethodGet && r.ContentLength > 0 {
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					logger.GetLogger().Error("Decode error", "error", err)
					respondWithError(w, NewBadRequestError(err.Error()))
					return
				}
			} else {
				// Parse query parameters
				queryValues := r.URL.Query()
				if queryValues != nil && len(queryValues) > 0 {
					for i := 0; i < reqType.NumField(); i++ {
						field := reqType.Field(i)
						queryTag := field.Tag.Get("query")
						if queryTag == "" {
							continue
						}
						if value := queryValues.Get(queryTag); value != "" {
							fieldValue := reqValue.Field(i)
							if fieldValue.CanSet() {
								// Convert string value to appropriate type
								switch fieldValue.Kind() {
								case reflect.String:
									fieldValue.SetString(value)
								case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
									if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
										fieldValue.SetInt(intVal)
									}
								case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
									if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
										fieldValue.SetUint(uintVal)
									}
								case reflect.Float32, reflect.Float64:
									if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
										fieldValue.SetFloat(floatVal)
									}
								case reflect.Bool:
									if boolVal, err := strconv.ParseBool(value); err == nil {
										fieldValue.SetBool(boolVal)
									}
								}
							}
						}
					}
				}
			}
			if rctx := chi.RouteContext(ctx); rctx != nil {
				urlParams := make(map[string]string)
				for i, key := range rctx.URLParams.Keys {
					if key != "*" {
						urlParams[key] = rctx.URLParams.Values[i]
					}
				}

				// Map URL parameters to struct fields with uri tag
				for i := 0; i < reqType.NumField(); i++ {
					field := reqType.Field(i)
					uriTag := field.Tag.Get("uri")
					if uriTag != "" {
						// Extract the parameter name from the uri tag
						paramName := uriTag
						if value, ok := urlParams[paramName]; ok && value != "" {
							fieldValue := reqValue.Field(i)
							if fieldValue.CanSet() {
								fieldValue.SetString(value)
							}
						}
					}
				}

				// Also try to map URL parameters by field name for backward compatibility
				for paramName, value := range urlParams {
					field := reqValue.FieldByName(paramName)
					if field.IsValid() && field.CanSet() {
						field.SetString(value)
					}
				}
			}

			// Validate request
			if err := validate.Struct(req); err != nil {
				logger.GetLogger().Error("Validation error", "error", err)
				respondWithError(w, NewBadRequestError(err.Error()))
				return
			}

		}
		// Call handler
		resp, err := handler(ctx, req)
		if err != nil {
			logger.GetLogger().Error("Handler error", "error", err)
			if apiErr, ok := err.(*APIError); ok {
				respondWithError(w, apiErr)
			} else {
				respondWithError(w, NewBadRequestError(err.Error()))
			}
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.GetLogger().Error("Encode error", "error", err)
			respondWithError(w, NewBadRequestError(err.Error()))
			return
		}
	}
}

// Helper function to respond with an error
func respondWithError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

// RequireAuth creates middleware that requires authentication
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				respondWithError(w, NewUnauthorizedError("Authorization header is required"))
				return
			}

			// Extract token from "Bearer <token>"
			if len(token) < 7 || token[:7] != "Bearer " {
				respondWithError(w, NewUnauthorizedError("Invalid Authorization header"))
				return
			}
			token = token[7:]

			claims, err := utils.ValidateToken(token)
			if err != nil {
				respondWithError(w, NewUnauthorizedError(err.Error()))
				return
			}
			db := r.Context().Value("database").(*models.DB)
			userID, err := uuid.Parse(claims.UserID)
			user, err := db.GetUserByID(r.Context(), userID)
			if err != nil {
				respondWithError(w, NewUnauthorizedError(err.Error()))
				return
			}
			claims.IsSuperUser = user.IsSuperuser

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "claims", claims)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
