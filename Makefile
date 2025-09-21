.PHONY: build run

# ========= Vars definitions =========

app = geo-filt

# ========= Prepare commands =========

tidy:
	go mod tidy
	go clean

del:
	rm ./$(app)* || echo "file didn't exists"
	rm ./trace*  || echo "file didn't exists"

# ========= Compile commands =========

build:
	go build -o ./build/$(app) -v ./cmd/$(app)/main.go

run: del build
	./$(app)

.DEFAULT_GOAL := run
