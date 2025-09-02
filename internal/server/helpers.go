package server

import (
	errs "errors"
	"log/slog"

	"github.com/dmitrii/llm-gateway/internal/errors"
	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	var typedError errors.Error
	if errs.As(err, &typedError) {
	} else {
		typedError = errors.ErrInternal.WithDetails(err)
	}
	slog.Error("Failed to execute request", "status", typedError.Status, "message", typedError.Message, "details", typedError.Details)
	c.JSON(typedError.Status, typedError)
}
