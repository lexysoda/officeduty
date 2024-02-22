include .env

BINARY := officeduty

.PHONY: build run clean ngrok run-ngrok

build:
	go build -o ${BINARY} main.go

run: build
	go run .

clean:
	rm ./officeduty

ngrok: build
	ngrok http --domain=$(NGROK_URL) 1337 > /dev/null

run-ngrok:
	@$(MAKE) -j 2 ngrok run
