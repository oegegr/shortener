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
		go build -o bin ./...

.PHONY: run-with-db
run-with-db: build-shortener run-postgresql  
	        BASE_URL=http://127.0.0.1:8080 \
		SERVER_ADDRESS=127.0.0.1:8080 \
		DATABASE_DSN=postgres://admin:admin@172.28.1.1:5432/url?sslmode=disable \
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
	  -p 172.28.1.1:5432:5432 \
	  -v postgres-data:/var/lib/postgresql/data \
	  postgres:latest 
	sleep 5

