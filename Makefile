.PHONY: build install

BIN_DIR := bin
BINARY := MCPProbe
INSTALL_DIR := $(HOME)/bin

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BIN_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	ln -sfn $(INSTALL_DIR)/$(BINARY) $(INSTALL_DIR)/probe
