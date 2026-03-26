package common

import (
	"net/http"
	"reflect"

	"github.com/gofiber/fiber/v3"
)

type response struct {
	context fiber.Ctx
}

type ResultData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func NewResponse(ctx fiber.Ctx) *response {
	return &response{context: ctx}
}

func (r *response) Success() error {
	return r.write(http.StatusOK, RETURN_SUCCESS, "success", nil)
}

func (r *response) SuccessWithData(data any) error {
	return r.write(http.StatusOK, RETURN_SUCCESS, "success", normalizeData(data))
}

func (r *response) Error(data any) error {
	appErr := normalizeError(data)
	return r.write(appErr.GetHTTPStatus(), appErr.GetErrorCode(), appErr.GetMsg(), nil)
}

func (r *response) ErrorWithCode(data any, status int) error {
	appErr := normalizeError(data)
	return r.write(status, appErr.GetErrorCode(), appErr.GetMsg(), nil)
}

func (r *response) write(status, code int, message string, data any) error {
	return r.context.Status(status).JSON(ResultData{
		Code:    code,
		Message: message,
		Data:    data,
		TraceID: requestTraceID(r.context),
	})
}

func requestTraceID(ctx fiber.Ctx) string {
	if traceID := ctx.GetRespHeader("X-Request-ID"); traceID != "" {
		return traceID
	}
	if traceID := ctx.Get("X-Request-ID"); traceID != "" {
		return traceID
	}
	if traceID := ctx.Get("X-Trace-ID"); traceID != "" {
		return traceID
	}
	if traceID, ok := ctx.Locals("trace_id").(string); ok {
		return traceID
	}
	return ""
}

func normalizeData(data any) any {
	value := reflect.ValueOf(data)
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if value.IsValid() && value.Kind() == reflect.Struct {
		field := value.FieldByName("ID")
		if field.IsValid() && field.Kind() == reflect.Int64 && field.Int() == 0 {
			return nil
		}
	}

	return data
}

func normalizeError(data any) Error {
	switch value := data.(type) {
	case nil:
		return NewServiceError("request failed")
	case Error:
		return value
	case error:
		return NewServiceError(value.Error())
	case string:
		return NewServiceError(value)
	default:
		return NewServiceError("request failed")
	}
}
