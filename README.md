# Oathkeeper + Github auth demo

This example setup is composed of three services:

1. Oathkeeper in reverse proxy mode
2. [A service that exchanges tokens for user details](./token-exchange)
3. [A service that does things with user details](./user-test)

## Requirements

- go 1.16
- Docker Compose
- [Github OAuth Client ID](https://github.com/settings/applications/new)

## Running

```sh
make all
docker-compose up -d
OATHKEEPER_GH_DEMO_CLIENT_ID=<your github client ID> ./dist/cli-test
```

## Expected Output

```
cd user-test && go vet && go get && go build -o ./../dist/user-test
cd cli-test && go vet && go get && go build -o ./../dist/cli-test
cd token-exchange && go vet && go get && go build -o ./../dist/token-exchange
cd user-test && docker build -t local-user-test .
Successfully tagged local-user-test:latest
cd token-exchange && docker build -t local-token-exchange .
Successfully tagged local-token-exchange:latest

Creating network "oathkeeper-gh-demo_exposed" with the default driver
Creating network "oathkeeper-gh-demo_lan" with the default driver
Creating oathkeeper-gh-demo_token-exchange_1 ... done
Creating oathkeeper-gh-demo_oathkeeper_1     ... done
Creating oathkeeper-gh-demo_web_1            ... done

INFO[0000] OATHKEEPER_GH_DEMO_API_URL was not set, using default  default="http://localhost/"
WARN[0000] failed reading persisted access token, will regenerate  error="open access-token: no such file or directory"
INFO[0000] please enter this code :)                     code=1234-ABCD url="https://github.com/login/device"
WARN[0007] poll failed with expected error               error-code=authorization_pending error-description="The authorization request is still pending." error-uri="https://docs.github.com/developers/apps/authorizing-oauth-apps#error-codes-for-the-device-flow"
INFO[0014] retrieved new token!                          access-token=12345
INFO[0014] hit our API :)                                output="Hello user 1234\nYou logged in with github using the username ghost and the email ghost@github.com\n"
```