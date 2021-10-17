BINARY_NAME=gitea-sonarqube-bot

export GO111MODULE=on

help:
	@echo "Make Routines:"
	@echo " - build                            Build the bot"
	@echo " - run    	                       Start the bot"
	@echo " - clean                            Delete generated files"
	@echo " - test    	                       Run test suite"
	@echo " - coverage    	                   Run test suite and generates coverage report as HTML file"
	@echo " - dep                              Dependency maintenance (tidy, vendor, verify)"
	@echo " - vet                              Examine Go source code and reports suspicious parts"
	@echo " - fmt                              Format the Go code"
	@echo " - help    	                       Print this help"

build:
	GOARCH=amd64 GOOS=linux go build --mod=vendor -o ${BINARY_NAME} ./cmd/gitea-sonarqube-bot/

run:
	./${BINARY_NAME}

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f cover.out cover.html

test:
	go test -v ./...

coverage:
	go test -v -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

dep:
	go mod tidy
	go mod vendor
	go mod verify

vet:
	go vet ./...

fmt:
	go fmt ./...
