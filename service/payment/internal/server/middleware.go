package server

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Response 統一響應結構
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseEncoder 統一響應編碼器
func ResponseEncoder() khttp.EncodeResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		if v == nil {
			v = map[string]interface{}{}
		}

		resp := Response{
			Status:  200,
			Message: "success",
			Data:    v,
		}

		codec, _ := khttp.CodecForRequest(r, "Accept")
		data, err := codec.Marshal(resp)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		return err
	}
}

// ErrorEncoder 統一錯誤編碼器
func ErrorEncoder() khttp.EncodeErrorFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		se := errors.FromError(err)

		resp := Response{
			Status:  500,
			Message: se.Message,
		}

		// 根據錯誤碼設置 HTTP 狀態碼
		httpStatus := http.StatusInternalServerError
		if se.Code == 400 {
			httpStatus = http.StatusBadRequest
		} else if se.Code == 404 {
			httpStatus = http.StatusNotFound
		} else if se.Code == 401 {
			httpStatus = http.StatusUnauthorized
		} else if se.Code == 403 {
			httpStatus = http.StatusForbidden
		}

		codec, _ := khttp.CodecForRequest(r, "Accept")
		data, _ := codec.Marshal(resp)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		w.Write(data)
	}
}
