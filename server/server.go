package server

import (
	"context"
	"fmt"
)

type FcRequest struct {
	ClientIP string
	Path     string
}

type FcResponse struct {
	StatusCode int
	Body       string
}

func NewFcResponse(statusCode int, format string, args ...interface{}) *FcResponse {
	return &FcResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf(format, args...),
	}
}

type FcServer interface {
	Init()
	DDNS(ctx context.Context, req FcRequest) *FcResponse
}
