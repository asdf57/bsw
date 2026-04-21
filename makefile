.PHONY: build build-frontend connect-db clean

build:
	go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -d cmd/api,internal/http/handlers,internal/http/router,internal/models/api,internal/models/db,internal/models/mappers,internal/db,internal/currency --parseDependency --parseInternal
	docker compose up -d --build
	@printf '\nAPI: %s\nSwagger UI: %s\n\n' \
		'http://localhost:8080' \
		'http://localhost:8080/swagger/index.html'

build-frontend:
	docker compose build frontend

connect-db:
	docker exec -it db psql -U user -d bswdb -p 5433

clean:
	docker compose down -v
