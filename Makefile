# Generate protobuf files
.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/master/*.proto

# Build master
.PHONY: build-master
build-master:
	go build -o bin/master ./master

# Build worker  
.PHONY: build-worker
build-worker:
	go build -o bin/worker ./worker

# Build everything
.PHONY: build
build: proto
	mkdir -p bin
	go build -o bin/master ./master
	go build -o bin/worker ./worker

# Run master
.PHONY: run-master
run-master: 
	go run ./cmd/master/main.go

# Run worker
.PHONY: run-worker
run-worker: 
	go run ./cmd/worker/main.go

# Clean
.PHONY: clean
clean:
	rm -rf bin/
	rm -f proto/master/*.pb.go

# Install protoc tools
.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Help
.PHONY: help
help:
	@echo "Commands:"
	@echo "  make install     - Install protoc tools"
	@echo "  make proto       - Generate protobuf files"
	@echo "  make build       - Build everything"
	@echo "  make run-master  - Run master server"
	@echo "  make run-worker  - Run worker"
	@echo "  make clean       - Clean build files"