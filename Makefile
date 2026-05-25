.PHONY: up down logs test fmt vet migrate-up migrate-down run

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

DB_URL=postgres://insider:insider_secret@localhost:5432/insider_league?sslmode=disable

migrate-up:
	docker run --rm --network host -v $(CURDIR)/migrations:/migrations migrate/migrate \
		-path=/migrations -database "$(DB_URL)" up

migrate-down:
	docker run --rm --network host -v $(CURDIR)/migrations:/migrations migrate/migrate \
		-path=/migrations -database "$(DB_URL)" down -all

run:
	go run ./cmd/server
