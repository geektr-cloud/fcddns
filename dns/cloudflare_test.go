package dns

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestCloudflareOperator_Update(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Skip("skip test due to missing .env file")
	}

	domain := os.Getenv("TEST_CLOUDFLARE_DOMAIN")
	host := os.Getenv("TEST_CLOUDFLARE_HOST")
	if domain == "" || host == "" {
		t.Skip("skip test due to missing TEST_CLOUDFLARE_DOMAIN or TEST_CLOUDFLARE_HOST")
	}

	operator := &CloudflareOperator{}
	err = operator.Update(context.Background(), domain, host, "1.2.3.4")
	if err != nil {
		t.Errorf("CloudflareOperator.Update() error = %v", err)
	}
}
