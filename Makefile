.PHONY: verify
verify: 
	go fmt lint ./...

.PHONY: test-all 
test-all: go test ./...

.PHONY: build-shortener
build-shortener: 
	        rm -rf bin 
		mkdir -p bin
		chmod +x -R bin
		go build -o bin ./...

.PHONY: run-with-env 
run-with-env: build-shortener 
		BASE_URL=http://127.0.0.1 \
		SERVER_ADDRESS=127.0.0.1:8081 \
		bin/shortener

.PHONY: run-with-flags 
run-with-flags: build-shortener 
		bin/shortener -a 127.0.0.1:8080 -b http://127.0.0.1
