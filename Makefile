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

.PHONY: run-with-env 
run-with-env: build-shortener run-postgresql  
	        BASE_URL=http://127.0.0.1:38511 \
		SERVER_ADDRESS=127.0.0.1:38511 \
		DATABASE_DSN=postgres://admin:admin@172.28.1.1:5432/shortener \
		bin/shortener

.PHONY: run-with-flags 
run-with-flags: build-shortener 
		bin/shortener -a 127.0.0.1:8080 -b http://127.0.0.1

.PHONY: test-integration
test-integration: build-shortener 
		"bin/shortenertest -test.v -test.run=^TestIteration$(iter)\$ -binary-path=bin/shortener"

.PHONY: run-postgresql
run-postgresql: 
	docker rm -f $$(docker ps -q  -f=name=postgres)
	docker run -d --name postgres   -e POSTGRES_USER=admin   -e POSTGRES_PASSWORD=admin   -e POSTGRES_DB=shortener   -p 172.28.1.1:5432:5432   postgres:latest
	sleep 5
