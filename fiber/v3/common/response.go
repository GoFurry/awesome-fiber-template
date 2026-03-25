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
	Code int `json:"code"`
	Data any `json:"data"`
}

func NewResponse(ctx fiber.Ctx) *response {
	return &response{
		context: ctx,
	}
}

func (r *response) Success() error {
	return r.context.Status(http.StatusOK).JSON(ResultData{
		Code: RETURN_SUCCESS,
		Data: "操作成功",
	})
}

func (r *response) SuccessWithData(data interface{}) error {
	value := reflect.ValueOf(data)
	if value.IsValid() {
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				data = nil
			} else {
				value = value.Elem()
			}
		}

		if value.IsValid() && value.Kind() == reflect.Struct {
			field := value.FieldByName("ID")
			if field.IsValid() && field.Kind() == reflect.Int64 && field.Int() == 0 {
				data = nil
			}
		}
	}

	return r.context.JSON(ResultData{
		Code: RETURN_SUCCESS,
		Data: data,
	})
}

func (r *response) Error(data interface{}) error {
	result := ResultData{
		Code: RETURN_FAILED,
		Data: data,
	}
	return r.context.JSON(result)
}

func (r *response) ErrorWithCode(data interface{}, code int) error {
	result := ResultData{
		Code: RETURN_FAILED,
		Data: data,
	}
	return r.context.Status(code).JSON(result)
}
