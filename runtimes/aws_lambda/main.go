package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/geektr-cloud/fcddns/server"
)

func main() {
	// print all envs
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); !ok {
		log.Fatal("AWS_LAMBDA_FUNCTION_NAME not found")
	}

	bs := &server.JwtServer{}
	bs.Init()

	lambda.Start(func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
		if strings.HasPrefix(event.RawPath, "/ddns/v1") {
			resp := bs.DDNS(ctx, server.FcRequest{
				Path:     event.RawPath,
				ClientIP: event.RequestContext.HTTP.SourceIP,
			})

			return &events.APIGatewayV2HTTPResponse{
				StatusCode: resp.StatusCode,
				Body:       resp.Body,
			}, nil
		}

		return nil, nil
	})
}
