.PHONY: help status infra-up infra-ps go-cache app-ps app-stop health \
	catalog-run cart-run order-run inventory-run payment-run notification-run bff-run bff-test \
	storefront-run storefront-build storefront-test \
	compose-local-build compose-local-up compose-local-ps compose-local-logs compose-local-smoke compose-local-down \
	ci-go ci-frontend ci-compose ci-local

GO_IMAGE := golang:1.26.6
DOCKER_NETWORK := platform-lab-network
WORKSPACE := $(CURDIR)
GO_VOLUMES := -v "$(WORKSPACE):/workspace" -v platform-go-mod:/go/pkg/mod -v platform-go-build:/root/.cache/go-build

help:
	@echo "platform Docker development commands"
	@echo ""
	@echo "  make infra-up         Start PostgreSQL, Redis, and Kafka"
	@echo "  make infra-ps         Show infrastructure status"
	@echo "  make catalog-run      Run Catalog on :8081"
	@echo "  make cart-run         Run Cart on :8082"
	@echo "  make order-run        Run Order on :8083"
	@echo "  make inventory-run    Run Inventory on :8084"
	@echo "  make payment-run      Run Payment on :8085"
	@echo "  make notification-run Run Notification on :8086"
	@echo "  make bff-run          Run Web BFF on :8080"
	@echo "  make bff-test         Test Web BFF"
	@echo "  make storefront-run   Run Angular storefront on :4200"
	@echo "  make storefront-build Build Angular storefront"
	@echo "  make storefront-test  Test Angular storefront once"
	@echo "  make compose-local-up  Build and start the complete stack"
	@echo "  make compose-local-ps  Show complete stack status"
	@echo "  make compose-local-smoke Test both Saga paths"
	@echo "  make compose-local-down Stop the complete stack"
	@echo "  make health           Check ports 8080-8086"
	@echo "  make app-ps           Show application containers"
	@echo "  make app-stop         Stop only application containers"
	@echo "  make ci-local         Run local CI validation"
	@echo "  make status           Show Git status"

status:
	git status --short

infra-up:
	docker compose -f platform/docker/compose.yaml up -d

infra-ps:
	docker compose -f platform/docker/compose.yaml ps

go-cache:
	docker volume create platform-go-mod
	docker volume create platform-go-build

catalog-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-catalog-service --network $(DOCKER_NETWORK) -p 8081:8081 \
	  -e HTTP_ADDR=:8081 -e DATABASE_URL=postgres://platform:platform@platform-lab-postgres:5432/catalog_db?sslmode=disable \
	  $(GO_VOLUMES) -w /workspace/services/catalog-service $(GO_IMAGE) go run ./cmd/api

cart-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-cart-service --network $(DOCKER_NETWORK) -p 8082:8082 \
	  -e HTTP_ADDR=:8082 -e REDIS_ADDR=platform-lab-redis:6379 \
	  $(GO_VOLUMES) -w /workspace/services/cart-service $(GO_IMAGE) go run ./cmd/api

order-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-order-service --network $(DOCKER_NETWORK) -p 8083:8083 \
	  -e HTTP_ADDR=:8083 -e DATABASE_URL=postgres://platform:platform@platform-lab-postgres:5432/orders_db?sslmode=disable \
	  -e KAFKA_BROKERS=platform-lab-kafka:29092 -e ORDERS_TOPIC=orders -e PAYMENTS_TOPIC=payments \
	  -e INVENTORY_TOPIC=inventory -e PAYMENTS_CONSUMER_GROUP=order-service-payments \
	  -e INVENTORY_CONSUMER_GROUP=order-service-inventory \
	  $(GO_VOLUMES) -w /workspace/services/order-service $(GO_IMAGE) go run ./cmd/api

inventory-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-inventory-service --network $(DOCKER_NETWORK) -p 8084:8084 \
	  -e HTTP_ADDR=:8084 -e DATABASE_URL=postgres://platform:platform@platform-lab-postgres:5432/inventory_db?sslmode=disable \
	  -e KAFKA_BROKERS=platform-lab-kafka:29092 -e ORDERS_TOPIC=orders -e INVENTORY_TOPIC=inventory \
	  -e KAFKA_CONSUMER_GROUP=inventory-service \
	  $(GO_VOLUMES) -w /workspace/services/inventory-service $(GO_IMAGE) go run ./cmd/worker

payment-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-payment-service --network $(DOCKER_NETWORK) -p 8085:8085 \
	  -e HTTP_ADDR=:8085 -e DATABASE_URL=postgres://platform:platform@platform-lab-postgres:5432/payments_db?sslmode=disable \
	  -e KAFKA_BROKERS=platform-lab-kafka:29092 -e INVENTORY_TOPIC=inventory -e PAYMENTS_TOPIC=payments \
	  -e KAFKA_CONSUMER_GROUP=payment-service -e PAYMENT_MAX_AUTH_CENTS=500000 \
	  $(GO_VOLUMES) -w /workspace/services/payment-service $(GO_IMAGE) go run ./cmd/worker

notification-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-notification-service --network $(DOCKER_NETWORK) -p 8086:8086 \
	  -e HTTP_ADDR=:8086 -e DATABASE_URL=postgres://platform:platform@platform-lab-postgres:5432/notifications_db?sslmode=disable \
	  -e KAFKA_BROKERS=platform-lab-kafka:29092 -e ORDERS_TOPIC=orders -e KAFKA_CONSUMER_GROUP=notification-service-v1 \
	  $(GO_VOLUMES) -w /workspace/services/notification-service $(GO_IMAGE) go run ./cmd/worker

bff-run: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm --name platform-lab-web-bff --network $(DOCKER_NETWORK) -p 8080:8080 \
	  -e HTTP_ADDR=:8080 -e CATALOG_SERVICE_URL=http://platform-lab-catalog-service:8081 \
	  -e CART_SERVICE_URL=http://platform-lab-cart-service:8082 -e ORDER_SERVICE_URL=http://platform-lab-order-service:8083 \
	  -e ALLOWED_ORIGIN=http://localhost:4200 \
	  $(GO_VOLUMES) -w /workspace/services/web-bff $(GO_IMAGE) go run ./cmd/api

bff-test: go-cache
	MSYS_NO_PATHCONV=1 docker run --rm -e GOWORK=off $(GO_VOLUMES) \
	  -w /workspace/services/web-bff $(GO_IMAGE) go test ./...

storefront-run:
	cd apps/storefront && npm start

storefront-build:
	cd apps/storefront && npm run build

storefront-test:
	cd apps/storefront && npm test -- --watch=false

compose-local-build:
	docker compose -f platform/docker/compose.yaml build

compose-local-up:
	docker compose -f platform/docker/compose.yaml up -d --build --wait

compose-local-ps:
	docker compose -f platform/docker/compose.yaml ps

compose-local-logs:
	docker compose -f platform/docker/compose.yaml logs -f --tail=100

compose-local-smoke:
	sh scripts/smoke-local.sh

compose-local-down:
	docker compose -f platform/docker/compose.yaml down

app-ps:
	docker ps --filter ancestor=$(GO_IMAGE) --format "table {{.Names}}\t{{.Ports}}\t{{.Status}}"

app-stop:
	-docker stop platform-lab-web-bff
	-docker stop platform-lab-notification-service
	-docker stop platform-lab-payment-service
	-docker stop platform-lab-inventory-service
	-docker stop platform-lab-order-service
	-docker stop platform-lab-cart-service
	-docker stop platform-lab-catalog-service

health:
	@for port in 8080 8081 8082 8083 8084 8085 8086; do \
	  printf "port %s: " "$$port"; curl -fsS "http://localhost:$$port/healthz" || true; echo; \
	done


CI_SERVICES := catalog-service cart-service order-service inventory-service payment-service notification-service web-bff

ci-go: go-cache
	@for service in $(CI_SERVICES); do \
	  echo "Testing $$service"; \
	  MSYS_NO_PATHCONV=1 docker run --rm -e GOWORK=off $(GO_VOLUMES) \
	    -w /workspace/services/$$service $(GO_IMAGE) go test ./... || exit 1; \
	done

ci-frontend:
	cd apps/storefront && npm ci
	cd apps/storefront && npm test -- --watch=false
	cd apps/storefront && npm run build

ci-compose:
	docker compose -f platform/docker/compose.yaml -f platform/docker/compose.observability.yaml config --quiet

ci-local: ci-go ci-frontend ci-compose
	@echo "platform local CI passed"
