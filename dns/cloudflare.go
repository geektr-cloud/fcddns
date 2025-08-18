package dns

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflareOperator struct {
	client atomic.Pointer[cloudflare.API]
}

func (o *CloudflareOperator) getClient() (*cloudflare.API, error) {
	if client := o.client.Load(); client != nil {
		return client, nil
	}

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN is not set")
	}

	client, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudflare client: %w", err)
	}

	o.client.Store(client)
	return client, nil
}

func (o *CloudflareOperator) Update(ctx context.Context, domain string, host string, ip string) error {
	client, err := o.getClient()
	if err != nil {
		return err
	}

	zoneID, err := client.ZoneIDByName(domain)
	if err != nil {
		return fmt.Errorf("failed to get zone id for domain %s: %w", domain, err)
	}

	fqdn := fmt.Sprintf("%s.%s", host, domain)
	records, _, err := client.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: fqdn,
	})
	if err != nil {
		return fmt.Errorf("failed to list A records for %s: %w", fqdn, err)
	}

	if len(records) > 1 {
		return fmt.Errorf("multiple A records found for %s", fqdn)
	}

	if len(records) == 0 {
		_, err := client.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
			Type:    "A",
			Name:    fqdn,
			Content: ip,
			TTL:     1, // 1 for auto
			Proxied: new(bool),
		})
		if err != nil {
			return fmt.Errorf("failed to create A record for %s: %w", fqdn, err)
		}
		return nil
	}

	record := records[0]
	if record.Content == ip {
		return nil
	}

	_, err = client.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    "A",
		Name:    fqdn,
		Content: ip,
	})
	if err != nil {
		return fmt.Errorf("failed to update A record for %s: %w", fqdn, err)
	}

	return nil
}

func init() {
	operator := &CloudflareOperator{}
	domains := os.Getenv("CLOUDFLARE_DOMAINS")
	for domain := range strings.SplitSeq(domains, ",") {
		SetOperator(domain, operator)
	}
}
