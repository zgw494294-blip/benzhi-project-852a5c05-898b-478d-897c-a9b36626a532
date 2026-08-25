package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

const maxRequestBody = 1 << 20

type protocolError struct {
	Status  int
	Code    string
	Message string
}

func (e *protocolError) Error() string { return e.Code + ": " + e.Message }

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &protocolError{Status: http.StatusUnsupportedMediaType, Code: "UNSUPPORTED_MEDIA_TYPE", Message: "Content-Type 必须为 application/json"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return &protocolError{Status: http.StatusRequestEntityTooLarge, Code: "REQUEST_TOO_LARGE", Message: "请求体不得超过 1 MiB"}
		}
		message := "请求 JSON 无效"
		if strings.Contains(err.Error(), "unknown field") {
			message = "请求包含未知 JSON 字段"
		}
		return &protocolError{Status: http.StatusBadRequest, Code: "INVALID_JSON", Message: message}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return &protocolError{Status: http.StatusBadRequest, Code: "INVALID_JSON", Message: "请求体只能包含一个 JSON 值"}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

type writeMeta struct {
	ExpectedVersion uint64 `json:"expectedVersion"`
	IdempotencyKey  string `json:"idempotencyKey"`
}
