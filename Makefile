main_package = ./
build_dir = ./bin

.PHONY: build
build:
	@mkdir -p ${build_dir}
	go build -o ${build_dir}/app ${main_package}

.PHONY: run
run:
	${build_dir}/app

.PHONY: dev
dev: build run

.PHONY: clean
clean: confirm
	@echo "Cleaning build artifacts..."
	@rm -rf ${build_dir}

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N]: " && read ans && [ $${ans:-N} = y ]
