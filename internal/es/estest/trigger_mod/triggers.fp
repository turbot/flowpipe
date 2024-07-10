trigger "schedule" "report_trigger" {
  schedule = "* * * * *"

  param "param_one" {
    type = string
    default = "value_one"
  }

  param "param_two" {
    type = string
    default = "value_two"
  }

  param "param_three" {
    type = number
    default = 42
  }

  param "param_four" {
    type = map(string)
    default = {
      "foo": "bar"
      "bar": "baz"
    }
  }

  param "param_five" {
    type = map(number)
    default = {
      "foo": 1
      "bar": 2
    }
  }

  pipeline = pipeline.github_issue

  args = {
    gh_repo = param.param_one
  }
}


pipeline "github_issue" {
  param "gh_repo" {
    type    = string
    default = "foo"
  }

  step "http" "get_issue" {
    url = "https://api.github.com/repos/octocat/${param.gh_repo}/issues/2743"
  }
}
