INSTALL_PATH = /etc/telegraf/plugins/inputs/execd/query

.PHONY: help
help : Makefile
	@echo "usage: make"
	@echo 
	@echo "TARGETS:"
	@sed -n 's/^## HELP://p' $<
	@echo 

## HELP:  all          Build all
.PHONY: all
all: build
	@echo 
	@echo ALL completed

## HELP:  clean        Clean all
.PHONY: clean
clean:
	@rm -fR ./dist
	@echo 
	@echo CLEAN completed

build:
	@go build -o ./dist/query cmd/main.go
	@cp "cmd/plugin.conf" ./dist

## HELP:  install      Install plugin
.PHONY: install
install: build install-config install-plugin
	@echo 
	@echo INSTALL completed

install-config:
	@mkdir -p ${INSTALL_PATH};
	@cp "./dist/plugin.conf" ${INSTALL_PATH};
	@echo "${INSTALL_PATH}/plugin.conf"

install-plugin:
	@mkdir -p ${INSTALL_PATH};
	@cp "./dist/query" ${INSTALL_PATH};
	@echo "${INSTALL_PATH}/query"

## HELP:  run          Run plugin
.PHONY: run
run: build
run: 
	@./dist/query --config ./dist/plugin.conf

## HELP:  test         Run test
.PHONY: test
test: test-plugin

test-plugin:
	go test ./plugins/inputs/query/ -v
