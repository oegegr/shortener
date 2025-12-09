.PHONY: verify
verify: 
	go fmt ./...
	golangci-lint run ./...


.PHONY: test-all 
test-all: 
	go test ./...

.PHONY: build-shortener
build-shortener: 
	        rm -rf bin 
		mkdir -p bin
		chmod +x -R bin
		go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$$(date +'%Y-%m-%d_%H:%M:%S') -X main.buildCommit=$$(git rev-parse HEAD)" -o bin ./...

.PHONY: run-with-db
run-with-db: build-shortener run-postgresql cert-clean cert
		CONFIG=config.json \
		TLS_CERT_FILE=cert.pem \
		TLS_KEY_FILE=key.pem \
		ENABLE_HTTPS=true \
	    BASE_URL=https://127.0.0.1:8080 \
		SERVER_ADDRESS=127.0.0.1:8080 \
		DATABASE_DSN=postgres://admin:admin@127.0.0.1:5432/url?sslmode=disable \
		AUDIT_FILE=/tmp/foobar \
		bin/shortener

.PHONY: run-with-dbfile 
run-with-dbfile: build-shortener 
	        BASE_URL=http://127.0.0.1:8080 \
		SERVER_ADDRESS=127.0.0.1:8080 \
		FILE_STORAGE_PATH=/tmp/foo \
		bin/shortener

.PHONY: run-with-mem
run-with-mem: build-shortener 
	        BASE_URL=http://127.0.0.1:8080 \
		SERVER_ADDRESS=127.0.0.1:8080 \
		bin/shortener

.PHONY: run-default
run-default: build-shortener 
		bin/shortener

.PHONY: test-integration
test-integration: build-shortener 
		"bin/shortenertest -test.v -test.run=^TestIteration$(iter)\$ -binary-path=bin/shortener"

.PHONY: run-postgresql
run-postgresql: 
	docker rm -f $$(docker ps -q  -f=name=postgres) || true
	docker run -d --name postgres \
	  -e POSTGRES_USER=admin \
	  -e POSTGRES_PASSWORD=admin \
	  -e POSTGRES_DB=url \
	  -p 127.0.0.1:5432:5432 \
	  -v postgres-data:/var/lib/postgresql/data \
	  postgres:17
	sleep 5

# Generate development TLS certificate
.PHONY: cert
cert:
	@echo "Generating development TLS certificate for localhost..."
	openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Development/OU=Dev/CN=localhost"

# Show certificate info
.PHONY: cert-info
cert-info:
	@echo "Certificate information:"
	openssl x509 -in cert.pem -text -noout
	@echo "\nPrivate key information:"
	openssl rsa -in key.pem -check -noout

# Clean certificate files
.PHONY: cert-clean
cert-clean:
	@echo "Cleaning up certificate files..."
	rm -f cert.pem key.pem

.PHONY: generate clean install-tools

PROTO_FILE = api/shortener.proto
GO_OUT = api
GOPATH_BIN = $(shell go env GOPATH)/bin

# Установить инструменты protoc
install-tools:
	@echo "Installing protoc tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed to $(GOPATH_BIN)"

# Проверить наличие инструментов
check-tools:
	@if [ ! -f "$(GOPATH_BIN)/protoc-gen-go" ]; then \
		echo "protoc-gen-go not found. Run 'make install-tools' first."; \
		exit 1; \
	fi
	@if [ ! -f "$(GOPATH_BIN)/protoc-gen-go-grpc" ]; then \
		echo "protoc-gen-go-grpc not found. Run 'make install-tools' first."; \
		exit 1; \
	fi
	@echo "All tools are available."

# Скачать proto файлы
download-protos:
	@echo "Downloading required proto files..."
	@mkdir -p third_party/google/protobuf
	@if [ ! -f "third_party/google/protobuf/empty.proto" ]; then \
		curl -s -o third_party/google/protobuf/empty.proto https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/empty.proto; \
		echo "Downloaded empty.proto"; \
	else \
		echo "empty.proto already exists"; \
	fi

# Генерация кода
generate: install-tools check-tools download-protos
	@echo "Generating Go code from protobuf..."
	@PATH="$(GOPATH_BIN):$$PATH" protoc \
		--go_out=$(GO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
		-I . -I ./third_party \
		$(PROTO_FILE)
	@echo "Code generation completed!"

clean:
	rm -rf $(GO_OUT)/*.pb.go
	rm -rf third_party/