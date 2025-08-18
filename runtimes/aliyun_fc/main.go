package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aliyun/fc-runtime-go-sdk/events"
	"github.com/aliyun/fc-runtime-go-sdk/fc"
	"github.com/geektr-cloud/fcddns/server"
)

func main() {
	if _, ok := os.LookupEnv("FC_RUNTIME_API"); !ok {
		log.Fatal("FC_RUNTIME_API not found")
	}

	bs := &server.JwtServer{}
	bs.Init()

	fc.Start(func(ctx context.Context, event events.HTTPTriggerEvent) (*events.HTTPTriggerResponse, error) {
		if event.Body == nil {
			return &events.HTTPTriggerResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf("the request did not come from an HTTP Trigger, event: %v", event),
			}, nil
		}

		if strings.HasPrefix(*event.RawPath, "/ddns/v1") {
			resp := bs.DDNS(ctx, server.FcRequest{
				Path:     *event.RawPath,
				ClientIP: *event.TriggerContext.Http.SourceIp,
			})

			return &events.HTTPTriggerResponse{
				StatusCode: resp.StatusCode,
				Body:       resp.Body,
			}, nil
		}

		return &events.HTTPTriggerResponse{
			StatusCode: http.StatusNotFound,
			Body:       fmt.Sprintf("path not found: %v", event.RawPath),
		}, nil
	})
}
