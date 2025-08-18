package dns

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
	"github.com/aliyun/fc-runtime-go-sdk/fccontext"
)

type AliyunOperator struct {
	client atomic.Pointer[alidns.Client]
}

type AliRecord = alidns.DescribeSubDomainRecordsResponseBodyDomainRecordsRecord

func (o *AliyunOperator) getClient(ctx context.Context) (*alidns.Client, error) {
	// if not in fc environment, reuse client
	if _, ok := fccontext.FromContext(ctx); !ok {
		if client := o.client.Load(); client != nil {
			return client, nil
		}
	}

	credential, err := credentials.NewCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	if fcCtx, ok := fccontext.FromContext(ctx); ok {
		if fcCtx.Credentials.AccessKeyId == "" {
			return nil, fmt.Errorf("credentials not found in fc context")
		}

		credential, err = credentials.NewCredential(&credentials.Config{
			Type:            tea.String("sts"),
			AccessKeyId:     tea.String(fcCtx.Credentials.AccessKeyId),
			AccessKeySecret: tea.String(fcCtx.Credentials.AccessKeySecret),
			SecurityToken:   tea.String(fcCtx.Credentials.SecurityToken),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create credential from fc context: %w", err)
		}
	}

	config := &openapi.Config{Credential: credential}
	config.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	client, err := alidns.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	o.client.Store(client)
	return client, nil
}

func (o *AliyunOperator) getARecords(client *alidns.Client, domain string, host string) ([]*AliRecord, error) {
	describeSubDomainRecordsRequest := &alidns.DescribeSubDomainRecordsRequest{
		SubDomain:  tea.String(fmt.Sprintf("%s.%s", host, domain)),
		DomainName: tea.String(domain),
	}

	runtime := &util.RuntimeOptions{}
	response, err := client.DescribeSubDomainRecordsWithOptions(describeSubDomainRecordsRequest, runtime)
	if err != nil {
		return nil, err
	}

	result := []*AliRecord{}

	for _, record := range response.Body.DomainRecords.Record {
		if record.Type == nil || *record.Type != "A" {
			continue
		}
		if record.Status == nil || *record.Status != "ENABLE" {
			continue
		}
		result = append(result, record)
	}

	return result, nil
}

func (o *AliyunOperator) createRecord(client *alidns.Client, domain string, host string, ip string) error {
	createDomainRecordRequest := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domain),
		RR:         tea.String(host),
		Type:       tea.String("A"),
		Value:      tea.String(ip),
	}

	runtime := &util.RuntimeOptions{}
	_, err := client.AddDomainRecordWithOptions(createDomainRecordRequest, runtime)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

func (o *AliyunOperator) updateRecord(client *alidns.Client, record *AliRecord, ip string) error {
	updateDomainRecordRequest := &alidns.UpdateDomainRecordRequest{
		RecordId: record.RecordId,
		RR:       record.RR,
		Type:     record.Type,
		Value:    tea.String(ip),
	}

	runtime := &util.RuntimeOptions{}
	_, err := client.UpdateDomainRecordWithOptions(updateDomainRecordRequest, runtime)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

func (o *AliyunOperator) Update(ctx context.Context, domain string, host string, ip string) error {
	client, err := o.getClient(ctx)
	if err != nil {
		return err
	}

	records, err := o.getARecords(client, domain, host)
	if err != nil {
		return fmt.Errorf("failed to get A records: %w", err)
	}

	if len(records) > 1 {
		return fmt.Errorf("multiple A records found for %s.%s", host, domain)
	}

	if len(records) == 0 {
		return o.createRecord(client, domain, host, ip)
	}

	if *records[0].Value == ip {
		return nil
	}

	return o.updateRecord(client, records[0], ip)
}

func init() {
	operator := &AliyunOperator{}
	for domain := range strings.SplitSeq(os.Getenv("ALIYUN_DOMAINS"), ",") {
		SetOperator(domain, operator)
	}
}
