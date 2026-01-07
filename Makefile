.PHONY: build run docker-build docker-push docker-run clean

# Variables
BINARY_NAME=orgmdns
BINARY_PATH=./$(BINARY_NAME)
IMAGE=orgmcr.or-gm.co/osmargm1202/orgmdns:latest
DOCKER_REGISTRY=orgmcr.or-gm.co

# Build local
build:
	@echo "Construyendo binario local..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/orgmdns
	@echo "Binario construido en $(BINARY_PATH)"

# Run local
run:
	@echo "Ejecutando aplicación localmente..."
	@go run ./cmd/orgmdns

# Run con debug
run-debug:
	@echo "Ejecutando aplicación localmente con debug..."
	@go run ./cmd/orgmdns --debug

# Docker build
docker-build:
	@echo "Construyendo imagen Docker..."
	@podman build -t $(IMAGE) .
	@echo "Imagen construida: $(IMAGE)"

# Docker push
docker-push:
	@echo "Subiendo imagen a $(DOCKER_REGISTRY)..."
	@podman push $(IMAGE)
	@echo "Imagen subida exitosamente"

# Docker run local
docker-run:
	@echo "Ejecutando contenedor localmente..."
	@podman run --rm -it \
		--env-file .env \
		-v $(PWD)/logs:/app/logs \
		$(IMAGE)

# Docker run con debug
docker-run-debug:
	@echo "Ejecutando contenedor localmente con debug..."
	@podman run --rm -it \
		--env-file .env \
		-v $(PWD)/logs:/app/logs \
		$(IMAGE) --debug

# Limpiar binarios y logs
clean:
	@echo "Limpiando binarios y logs..."
	@rm -rf ./$(BINARY_NAME)
	@rm -rf logs/*.log
	@echo "Limpieza completada"

# Instalar dependencias
deps:
	@echo "Descargando dependencias..."
	@go mod download
	@go mod tidy

# Test (si se agregan tests en el futuro)
test:
	@echo "Ejecutando tests..."
	@go test ./...
