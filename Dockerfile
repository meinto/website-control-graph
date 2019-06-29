FROM golang:1.12-stretch AS builder

WORKDIR /go/src/github.com/meinto/gqlgen-starter
COPY . .

RUN GO111MODULE=on go get ./...
RUN GO111MODULE=on go generate ./...
RUN GO111MODULE=on GOOS=linux GOARCH=386 go build -o gql-server -ldflags "-X github.com/meinto/gqlgen-starter/chrome.AddExecPathInBuild=yes" server/server.go


FROM chromedp/headless-shell:74.0.3729.1

COPY --from=builder /go/src/github.com/meinto/gqlgen-starter/gql-server .

COPY docker.startup.sh .

EXPOSE 8080

ENTRYPOINT [ "/docker.startup.sh" ]
