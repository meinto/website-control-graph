# GqlGen Starter Project

This is a quickstart project for [gqlgen](https://github.com/99designs/gqlgen).

## Setup

```bash
git clone https://github.com/meinto/gqlgen-starter.git
cd gqlgen-starter
go run github.com/99designs/gqlgen
go run server/server.go
```

## Test it

Open localhost:8080 and copy-paste the query or mutation

```
query {
  hello(name: "Lars")
}
``` 

```
mutation {
  foo {
    bar
  }
}
```

## Start your project

The project structure is like follows:

```
/schema               # folder for all your schemas
  query.gql           # all your queries go here
  mutation.gql        # all your mutations go here
  other.gql           # place your types in seperate files

/resolver             # resolver package
  resolver.go         # root resolver
  other.go            # place other resolvers in seperate files

/model                # model package
  genereated.go       # all generated models (don't edit this file)
  customModel.go      # place all models which you want to define by your own in seperate model files
                      # don't forget to define this in the config: https://gqlgen.com/config/

/server
  server.go           # your server

/graph
  /generated
    generated.go      # generated graphql (don't edit this file)
```

1. Now write your types, queries & mutations,
2. generate the graphql: `go run github.com/99designs/gqlgen`, 
3. implement the resolvers
4. and start your server: `go run server/server.go`.
5. **Have fun! :)**

## Docker

Build your graphql as docker container:

```bash
docker build -t gql-server .
docker run -p 8080:8080 gql-server
```


