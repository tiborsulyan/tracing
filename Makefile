build:
	go build -tags netgo,osusergo -gcflags '-N -l' -v -o ./cmd/servicea ./cmd/servicea
	go build -tags netgo,osusergo -gcflags '-N -l' -v -o ./cmd/serviceb ./cmd/serviceb

build.docker: build
	docker compose build

start:
	docker compose up -d

stop:
	docker compose down

clean:
	rm -f ./cmd/servicea/servicea ./cmd/serviceb/serviceb