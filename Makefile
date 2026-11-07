APP=my-auth

.PHONY: build clean all

build:
	go build -o $(APP) .

all: clean
	GOOS=darwin GOARCH=arm64 go build -o dist/$(APP)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/$(APP)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o dist/$(APP)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o dist/$(APP)-windows-amd64.exe .

clean:
	rm -rf dist/ $(APP)
