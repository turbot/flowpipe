pipeline "github_issue" {
    step "http" "get_issues" {
        url = "https://api.github.com/repos/octocat/hello-world/issues"
    }
}