build:
	go build -o ./cmd/servicea ./cmd/servicea
	go build -o ./cmd/serviceb ./cmd/serviceb

build.docker: build
	docker compose build

start:
	docker compose up -d

stop:
	docker compose down
