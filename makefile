build:
	@go build -o th-note .

run: build
	@./th-note