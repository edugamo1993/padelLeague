# Nombre del binario
BINARY_NAME=ligapadel

# Nombre del módulo y ruta al paquete que contiene la variable Version
MODULE_PATH=github.com/tuusuario/ligapadel/internal/info

# Versión por defecto si no se usa Git
VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: build run clean

build:
	go build -ldflags="-X '$(MODULE_PATH).Version=$(VERSION)'" -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)