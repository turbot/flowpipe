
trigger "schedule" "report_triggers" {
  schedule = "* * * * *"
  pipeline = pipeline.github_issue
  args = {
    github_token = var.default_gh_repo
    slack_token = var.var_one
  }
}


trigger "schedule" "report_triggers_with_schedule_var_with_default_value" {
  schedule = var.schedule_default
  pipeline = pipeline.github_issue
  args = {
    github_token = var.default_gh_repo
    slack_token = var.var_one
  }
}

trigger "schedule" "report_triggers_with_interval_var_with_default_value" {
  schedule = var.interval_default
  pipeline = pipeline.github_issue
  args = {
    github_token = var.default_gh_repo
    slack_token = var.var_one
  }
}
