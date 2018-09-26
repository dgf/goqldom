const versionQuery = `{
  version
}
`;

const golangQuery = `{
  golang_blog_articles: get(url: "https://blog.golang.org/index") {
    statusCode
    statusMessage
    contentType
    document {
      location
      title
      articles: select(selector: ".blogtitle") {
        all: elements {
          title: text(selector: "a")
          date: text(selector: ".date")
          link: attr(selector: "a", key: "href")
        }
      }
    }
  }
}
`;

window.addEventListener('load', function () {
    GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: "https://api.graph.cool/simple/v1/swapi",
        workspaceName: "goqldom service",
        settings: {
            'general.betaUpdates': false,
            'editor.cursorShape': 'line', // possible values: 'line', 'block', 'underline'
            'editor.fontSize': 13,
            'editor.fontFamily': `'Source Code Pro', 'Consolas', 'Inconsolata', 'Droid Sans Mono', 'Monaco', monospace`,
            'editor.theme': 'light', // possible values: 'dark', 'light'
            'editor.reuseHeaders': true,
            'request.credentials': 'omit', // possible values: 'omit', 'include', 'same-origin'
            'tracing.hideTracingResponse': true,
        },
        tabs: [
            {
                name: "Service version",
                endpoint: "http://localhost:8080/graphql",
                query: versionQuery
            }, {
                name: "Golang blog articles",
                endpoint: "http://localhost:8080/graphql",
                query: golangQuery
            }
        ]
    })
});