clean:
	@rm webserver 2> /dev/null || true

lint-front:
	@npm install && npm run lint

build-front:
	npm run build
	rm -rf ./internal/webserver/handlers/home/css && cp -rf ./frontend/styles ./internal/webserver/handlers/home/css
	gzip -c ./frontend/index.html > ./internal/webserver/handlers/home/index.html.gz
	gzip -f ./internal/webserver/handlers/home/index.js
	gzip -f ./internal/webserver/handlers/home/index.js.map

lint-back:
	@golangci-lint run

test-back:
	@go test -v ./internal/...

build-back:
	@go build -ldflags "-s -w" -o app cmd/main.go

run: build-front
	@go run cmd/main.go
