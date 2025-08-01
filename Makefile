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
run-with-env: build-shortener 
	        BASE_URL=http://127.0.0.1:38511 \
		SERVER_ADDRESS=127.0.0.1:38511 \
		bin/shortener

.PHONY: run-with-flags 
run-with-flags: build-shortener 
		bin/shortener -a 127.0.0.1:8080 -b http://127.0.0.1

.PHONY: test-integration
test-integration: build-shortener 
		"bin/shortenertest -test.v -test.run=^TestIteration$(iter)\$ -binary-path=bin/shortener"
