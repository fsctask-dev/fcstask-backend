MODULE_NAME := fcstask

.PHONY: init tidy

init:
	@if [ ! -f go.mod ]; then \
		echo "Init repo: $(MODULE_NAME)"; \
		go mod init $(MODULE_NAME); \
	else \
		echo "good. already exists"; \
	fi

tidy:
	go mod tidy

gen:
	go generate ./...

test:
	go test ./... -v