FROM golang:1.12-stretch AS builder

WORKDIR /go/src/github.com/meinto/website-control-graph
COPY . .

RUN GO111MODULE=on go get github.com/99designs/gqlgen
RUN GO111MODULE=on go generate ./...
RUN GO111MODULE=on GOOS=linux GOARCH=386 go build -o gql-server -ldflags "-X github.com/meinto/website-control-graph/chrome.DockerBuild=yes" server/server.go


FROM chromedp/headless-shell:74.0.3729.1

COPY --from=builder /go/src/github.com/meinto/website-control-graph/gql-server .

COPY docker.startup.sh .

EXPOSE 4000

ENTRYPOINT [ "/docker.startup.sh" ]
