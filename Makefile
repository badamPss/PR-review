.PHONY: build run clean up down logs rebuild ps lint

BINARY=pr-review
CMD=./cmd/pr-review
CONFIG=./config/local/config.yaml

build:
	go build -o $(BINARY) $(CMD)

run: build
	./$(BINARY) -config $(CONFIG)

clean:
	go clean
	rm -f $(BINARY)

up:
	docker-compose up -d

rebuild:
	docker-compose up --build -d

down:
	docker-compose down

logs:
	docker-compose logs -f

ps:
	docker-compose ps

lint:
	@golangci-lint --version
	CGO_ENABLED=0 golangci-lint run -v
