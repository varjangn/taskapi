build:
	@go build -o bin/taskapi

run: build
	@./bin/taskapi

test:
	@go test -f ./ ...
