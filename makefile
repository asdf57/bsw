.PHONY: build build-frontend connect-db clean cloud-up cloud-down

build:
	go run github.com/swaggo/swag/cmd/swag@latest init
	docker compose up -d --build

build-frontend:
	docker compose build frontend

connect-db:
	docker exec -it db psql -U user -d bswdb -p 5433

clean:
	docker compose down -v

cloud-up:
	docker compose --env-file cloud/.env -f cloud/docker-compose.yml up -d --build

cloud-down:
	docker compose --env-file cloud/.env -f cloud/docker-compose.yml down -v
