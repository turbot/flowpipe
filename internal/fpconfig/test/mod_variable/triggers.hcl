
trigger "schedule" "report_triggers" {
  schedule = "* * * * *"
  pipeline = pipeline.github_issue
  args = {
    github_token = var.default_gh_repo
    slack_token = var.var_one
  }
}
