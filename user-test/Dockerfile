FROM golang:alpine AS builder
ENV CGO_ENABLED=0
COPY . /app
RUN cd /app && go get && go build -o /dist/user-test

FROM scratch
COPY --from=builder /dist/user-test /dist/user-test
ENTRYPOINT ["/dist/user-test"]
