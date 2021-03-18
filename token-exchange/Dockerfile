FROM golang:alpine AS builder
ENV CGO_ENABLED=0
COPY . /app
RUN cd /app && go get && go build -o /dist/token-exchange

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /dist/token-exchange /dist/token-exchange
ENTRYPOINT ["/dist/token-exchange"]
