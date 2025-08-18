package dns

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestAliyunOperator_Update(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Skip("skip test due to missing .env file")
	}

	domain := os.Getenv("TEST_ALIYUN_DOMAIN")
	host := os.Getenv("TEST_ALIYUN_HOST")
	if domain == "" || host == "" {
		t.Skip("skip test due to missing TEST_ALIYUN_DOMAIN or TEST_ALIYUN_HOST")
	}

	operator := &AliyunOperator{}
	err = operator.Update(context.Background(), domain, host, "1.2.3.4")
	if err != nil {
		t.Errorf("AliyunOperator.Update() error = %v", err)
	}
}
