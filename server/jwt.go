package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/geektr-cloud/fcddns/dns"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Domain string `json:"domain"`
	Host   string `json:"host"`
	IP     string `json:"ip"`
	jwt.RegisteredClaims
}

type JwtServer struct {
	secret string
}

func (s *JwtServer) Init() {
	s.secret = os.Getenv("JWT_SECRET")
}

func (s *JwtServer) DDNS(ctx context.Context, req FcRequest) *FcResponse {
	if !strings.HasPrefix(req.Path, "/ddns/v1") {
		return NewFcResponse(http.StatusBadRequest, "invalid path: %s", req.Path)
	}

	// req.Path: /ddns/v1/{jwt}[/ip]
	parts := strings.Split(strings.TrimPrefix(req.Path, "/ddns/v1/"), "/")
	if len(parts) < 1 {
		return NewFcResponse(http.StatusBadRequest, "invalid path: %s", req.Path)
	}

	token := parts[0]
	if token == "" {
		return NewFcResponse(http.StatusBadRequest, "invalid path: %s", req.Path)
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return NewFcResponse(http.StatusBadRequest, "failed to parse token: %v", err)
	}

	if claims.Domain == "" || claims.Host == "" {
		return NewFcResponse(http.StatusBadRequest, "invalid claims: host: %s, domain: %s", claims.Host, claims.Domain)
	}

	if len(parts) >= 2 {
		req.ClientIP = parts[1]
	} else if claims.IP != "" {
		req.ClientIP = claims.IP
	}

	if req.ClientIP == "" {
		return NewFcResponse(http.StatusBadRequest, "client ip not found")
	}

	fqdn := fmt.Sprintf("%s.%s", claims.Host, claims.Domain)
	fmt.Printf("DDNS: %s => %s\n", fqdn, req.ClientIP)

	operator := dns.GetOperator(claims.Domain)
	if operator == nil {
		return NewFcResponse(http.StatusBadRequest, "operator not found for domain: %s", claims.Domain)
	}

	if err := operator.Update(ctx, claims.Domain, claims.Host, req.ClientIP); err != nil {
		return NewFcResponse(http.StatusInternalServerError, "failed to update record: %v", err)
	}

	return NewFcResponse(http.StatusOK, "ok")
}
