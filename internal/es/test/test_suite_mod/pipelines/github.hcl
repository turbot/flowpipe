pipeline "github_issue" {
    step "http" "get_issues" {
        url = "https://api.github.com/repos/octocat/hello-world/issues"
    }
}

pipeline "github_get_issue" {

  param "github_token" {
    type = string
  }

  param "github_issue_number" {
    type = number
  }

  step "http" "get_issue" {
    title  = "Get details about an issue"
    method = "post"
    url    = "https://api.github.com/graphql"
    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.github_token}"
    }

    request_body = jsonencode({
      query = <<EOM
              query {
                repository(owner: "octocat", name: "hello-world") {
                  issue(number: ${param.github_issue_number}) {
                    id
                    number
                    url
                    title
                    body
                  }
                }
              }
            EOM
    })
  }

  output "issue" {
    value = step.http.get_issue.response_body
  }
}
