# Aliyun FC DDNS

## CLI Utils

```bash
go install
# go install github.com/geektr-cloud/fcddns@latest

# Sign the request
fcddns jwt sign example.com www

# Verify the request
fcddns jwt verify "..."
```

## FC Deploy

### Environment

```bash
JWT_SECRET=...

# aliyun dns
ALIBABA_CLOUD_ACCESS_KEY_ID=
ALIBABA_CLOUD_ACCESS_KEY_SECRET=
ALIBABA_CLOUD_SECURITY_TOKEN=

ALIYUN_DOMAINS=foo.com

# cloudflare dns
CLOUDFLARE_API_TOKEN=

CLOUDFLARE_DOMAINS=bar.com
```

### Aliyun FC

```bash
# Build
CGO_ENABLED=0 go build -o main ./runtimes/aliyun_fc && zip -r aliyun-fc.zip main && rm main

# Deploy
# https://help.aliyun.com/zh/functioncompute/fc-3-0/user-guide/compile-and-deploy-code-packages-in-a-go-runtime

# Test
curl https://test-xxxx.xx-xxxxxxxx.fcapp.run/ddns/v1/$(fcddns jwt sign foo.com www)[/ip]
```

### AWS Lambda

```bash
# Build
CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./runtimes/aws_lambda && zip -r aws-lambda-amd64.zip bootstrap && rm bootstrap
CGO_ENABLED=0 GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./runtimes/aws_lambda && zip -r aws-lambda-arm64.zip bootstrap && rm bootstrap

# Deploy
# https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html

# Test
curl https://test-xxxx.xx-xxxxxxxx.fcapp.run/ddns/v1/$(fcddns jwt sign foo.com www)[/ip]
```

### Cloudflare Workers

```bash
cd runtimes/cf-workers
pnpx wrangler dev
curl http://localhost:8787/ddns/v1/$(fcddns jwt sign anitya.net fctest.ddns)[/ip]

npx wrangler secret put JWT_SECRET
npx wrangler secret put CLOUDFLARE_API_TOKEN
sed -i 's/ddns.geektr.cloud/ddns.yourdomain.com/g' wrangler.jsonc
pnpx wrangler deploy
```
