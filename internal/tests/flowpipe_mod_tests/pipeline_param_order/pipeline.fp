pipeline "github_issue" {
  param "gh_repo" {
    type    = string
    default = "bar"
  }

  param "azure_repo" {
    type    = string
  }  

  param "gcp_repo" {
    type = string
    default = "foo"
  }

  step "http" "get_issue" {
    url = "https://api.github.com/repos/octocat/${param.gh_repo}/issues/2743"
  }
}

