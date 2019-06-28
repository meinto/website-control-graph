FROM golang:1.12-stretch AS builder

WORKDIR /go/src/github.com/meinto/gqlgen-starter
COPY . .

RUN GO111MODULE=on go get ./...
RUN GO111MODULE=on go generate ./...
RUN GO111MODULE=on GOOS=linux GOARCH=386 go build -o gql-server server/server.go


FROM golang:1.12-alpine

WORKDIR /app/
COPY --from=builder /go/src/github.com/meinto/gqlgen-starter/gql-server .

CMD /app/gql-server
