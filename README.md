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

Open http://localhost:4000 and copy-paste the query. You will get the `h1` and `h2` headlines of the [hello world wikipedia article](https://en.wikipedia.org/wiki/%22Hello,_World!%22_program) as result.

```
query {
  control(
    actions:[
      {navigate:"https://en.wikipedia.org/wiki/%22Hello,_World!%22_program"},
      {waitVisible:"h1"}
    ]
    output: [
      {
        name: "Überschriften"
        selectors: [
          {
            type: string_array
            cssSelector: "h1"
            key: "headline1"
          }
          {
            type: string_array
            cssSelector: "h2"
            key: "headline2"
          }
        ]
      }
    ]
  ) {
    output
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
  runtimeVar: Selector # save a string from the website to use it in a following action
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
        name: "Name of collection",
        selectors: [ ... ]
      }
    ]
  ) {
    output
  }
}
```

### Example runtime variable

The result of the following query will be the `h1` Headline of the [Hello_World_(disambiguation)](https://en.wikipedia.org/wiki/Hello_World_(disambiguation)) page. This is the first link of the [hello world wikipedia article](https://en.wikipedia.org/wiki/%22Hello,_World!%22_program).

```
query {
  control(
    actions:[
      {navigate:"https://en.wikipedia.org/wiki/%22Hello,_World!%22_program"},
      {waitVisible:"h1"}
      {runtimeVar: {                              # store the link to Hello_World_(disambiguation) page
        cssSelector: ".mw-disambig"
        HTMLAttribute: "href"
        key: "runtimevar"
        type: string_prop
      }}
      {navigate:"https://en.wikipedia.org$0"},    # use the link to Hello_World_(disambiguation) page ($0)
      {waitVisible:"h1"}
    ]
    output: [
      {
        name: "Überschriften"
        selectors: [
          {
            type: string_array
            cssSelector: "h1"
            key: "headline1"
          }
        ]
      }
    ]
  ) {
    runtimeVars {
      name
      value
    }
    output
  }
}
```

## Docker

Build your graphql as docker container:

```bash
docker build -t website-control-graph .
docker run -p 4000:4000 website-control-graph
```


