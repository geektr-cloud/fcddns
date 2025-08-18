package dns

import (
	"context"
	"strings"
)

type RecordOperator interface {
	Update(ctx context.Context, domain string, host string, ip string) error
}

var operators = map[string]RecordOperator{}

func GetOperator(domain string) RecordOperator {
	o := operators[domain]
	if o == nil {
		o = operators["*"]
	}

	return o
}

func SetOperator(domain string, operator RecordOperator) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return
	}

	operators[domain] = operator
}
