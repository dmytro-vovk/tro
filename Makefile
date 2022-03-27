clean:
	@rm app 2> /dev/null || true

lint-front:
	@npm install && npm run lint

build-front:
	scripts/build-front.sh

lint-back:
	@golangci-lint run

test-back:
	@go test -v ./internal/...

build-back:
	scripts/build-back.sh

run: build-front
	docker-compose up --detach --build
	go run cmd/main.go; docker-compose down
