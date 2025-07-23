#!/bin/bash

set -e

# 创建证书目录
CERT_DIR="/tmp/k8s-webhook-server/serving-certs"
mkdir -p $CERT_DIR

echo "Generating webhook certificates..."

# 生成私钥
openssl genrsa -out $CERT_DIR/tls.key 2048

# 生成证书签名请求
openssl req -new -key $CERT_DIR/tls.key -out $CERT_DIR/tls.csr -subj "/CN=webhook-service.default.svc/O=webhook-service"

# 生成自签名证书，包含多个DNS名称
openssl x509 -req -in $CERT_DIR/tls.csr -signkey $CERT_DIR/tls.key -out $CERT_DIR/tls.crt -days 365 -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = webhook-service
DNS.2 = webhook-service.default
DNS.3 = webhook-service.default.svc
DNS.4 = webhook-service.default.svc.cluster.local
DNS.5 = localhost
IP.1 = 127.0.0.1
EOF
)

# 清理临时文件
rm $CERT_DIR/tls.csr

echo "Webhook certificates generated successfully in $CERT_DIR"
echo "Certificate details:"
openssl x509 -in $CERT_DIR/tls.crt -text -noout | grep -A 5 "Subject Alternative Name" || echo "SAN not found"
echo ""
echo "You can now run the webhook with: make run"