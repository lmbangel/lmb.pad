.PHONY: clean

build:
	go build -o bin/main

run:build
	./bin/main

tidy:
	go mod tidy

clean: 
	rm ./bin/* | rm ./db/*.json