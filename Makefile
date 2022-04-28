PROJECT_ROOT=$(shell git rev-parse --show-toplevel)

default: build

build: ## Build xfuzz binary
	@bash ${PROJECT_ROOT}/script/build.bash

build_cases: ## Build cases binary
	@bash ${PROJECT_ROOT}/script/build_cases.bash
