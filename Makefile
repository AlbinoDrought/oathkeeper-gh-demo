dist/user-test: user-test/go.mod user-test/main.go
	cd user-test && go vet && go get && go build -o ./../dist/user-test

dist/cli-test: cli-test/go.mod cli-test/main.go
	cd cli-test && go vet && go get && go build -o ./../dist/cli-test

dist/token-exchange: token-exchange/go.mod token-exchange/main.go
	cd token-exchange && go vet && go get && go build -o ./../dist/token-exchange

.PHONY: containers
containers:
	cd user-test && docker build -t local-user-test .
	cd token-exchange && docker build -t local-token-exchange .

.PHONY: all
all: dist/user-test dist/cli-test dist/token-exchange containers
