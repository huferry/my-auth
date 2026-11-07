APP=my-auth

.PHONY: build tray clean all

build:
	go build -o $(APP) .

tray:
	go build -tags tray -o $(APP)-tray .

all: clean
	GOOS=darwin GOARCH=arm64 go build -o dist/$(APP)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/$(APP)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o dist/$(APP)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o dist/$(APP)-windows-amd64.exe .
	GOOS=darwin GOARCH=arm64 go build -tags tray -o dist/$(APP)-tray-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -tags tray -o dist/$(APP)-tray-darwin-amd64 .

clean:
	rm -rf dist/ $(APP) $(APP)-tray
