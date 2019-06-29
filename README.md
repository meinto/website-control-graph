# Website Control Graph

Website Control Graph is a webscraper which allows you to control websites via graphql.

## Setup

```bash
git clone https://github.com/meinto/website-control-graph.git
cd website-control-graph
go run github.com/99designs/gqlgen
go run server/server.go
```

## Test it

Open http://localhost:8080 and copy-paste the query. You will get the `h1` and `h2` headlines of the [hello world wikipedia article](https://en.wikipedia.org/wiki/%22Hello,_World!%22_program) as result.

```
query {
  control(
    actions:[
      {navigate:"https://en.wikipedia.org/wiki/%22Hello,_World!%22_program"},
      {waitVisible:"h1"}
    ]
  	output: [
      {
        selector: "h1"
        key: "h1-list"
      }
      {
        selector: "h2 .mw-headline"
        key: "h2-list"
      }
    ]
  ) {
    value
  }
}
```

## Actions

The action type consists of different properties which you can use as input for your query. All actions will be executed as a queue.

```
input Action {
  navigate: String     # navigate to url
  sleep: Int           # sleep n seconds
  waitVisible: String  # wait till a specific element is visible on page
  sendKeys: Input      # fill data into an input
  click: String        # click a specific element on a page
  evalJS: String       # execute javascript
}
```

### Example login

```
query {
  control(
    actions:[
      {navigate:"https://your-website-with-login.com/login"},
      {waitVisible:"input[name='user']"}
      {sendKeys:{
        selector: "input[name='user']"
        value:"your-name"
      }}
      {sendKeys:{
        selector: "input[name='password']"
        value:"your-pass"
      }}
      {click:"input[type='submit']"}
      {waitVisible:"p.content-you-want-to-query"}
    ]
  	output: [
      {
        selector: "p.content-you-want-to-query"
        key: "contentList"
      }
    ]
  ) {
    key
    value
    index
    selector
  }
}
```

## Docker

Build your graphql as docker container:

```bash
docker build -t website-control-graph .
docker run -p 8080:8080 website-control-graph
```


