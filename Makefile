IP = "91.196.151.105"

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

deploy: build-front build-back
	# Authenticate by key pair
	scp app ${IP}:/opt/tro/app.new
	# You must be in 'sudo' group for this to work
	ssh ${IP} 'sudo systemctl stop tro; sudo mv /opt/tro/app.new /opt/tro/app; sudo chown tro17.tro17 /opt/tro/app; sudo systemctl start tro'
