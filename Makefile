build:
	@go build -o bin/study_academy_golang cmd/main.go


test:
	@go test -v ./...


run: build
	@./bin/study_academy_golang

migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir cmd/migrate/migrations -seq $$name


migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down


