package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
)

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestID"`
	} `json:"error"`
}

func respondError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务处理请求失败"
	var protocol *protocolError
	var business *domain.Error
	if errors.As(err, &protocol) {
		status, code, message = protocol.Status, protocol.Code, protocol.Message
	} else if errors.As(err, &business) {
		code, message = business.Code, business.Message
		switch business.Code {
		case "NOT_FOUND":
			status = http.StatusNotFound
		case "STATE_CONFLICT":
			status = http.StatusConflict
		default:
			status = http.StatusUnprocessableEntity
		}
	} else if application.IsVersionConflict(err) || application.IsIdempotencyConflict(err) {
		status, code, message = http.StatusConflict, "CONCURRENCY_CONFLICT", err.Error()
	} else {
		logger.Error("请求处理失败", "requestID", requestIDFrom(r.Context()), "error", err)
	}
	response := errorResponse{}
	response.Error.Code, response.Error.Message, response.Error.RequestID = code, message, requestIDFrom(r.Context())
	writeJSON(w, status, response)
}
