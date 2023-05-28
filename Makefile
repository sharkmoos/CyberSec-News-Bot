GO := go
GO_BUILD := $(GO) build
GO_CLEAN := $(GO) clean
GO_TEST := $(GO) test

APP_NAME := comsecDiscordBot
SRC_DIR := ./src
BUILD_DIR := ./bin

build:
	$(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)/

clean:
	$(GO_CLEAN)
	rm -f $(BUILD_DIR) 

# Test target
# test:
#	$(GO_TEST) ./$(SRC_DIR)/...

# Default target
.DEFAULT_GOAL := build

