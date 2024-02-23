pipeline "list_issues_using_query" {

  param "github_repo_full_name" {
    type = string
    default = "turbot/steampipe"
  }

  step "query" "list_issues" {
    database = "postgres://steampipe@host.docker.internal:9193/steampipe"
    sql      = "select number,url,title,body from github.github_issue where repository_full_name ='turbot/steampipe' limit 10"
  }

  output "val" {
    value = step.query.list_issues.rows
  }
}

pipeline "list_issues_using_query_param" {

  param "github_repo_full_name" {
    type = string
    default = "turbot/steampipe"
  }

  step "query" "list_issues" {
    database = "postgres://steampipe@host.docker.internal:9193/steampipe"
    sql      = "select number,url,title,body from github.github_issue where repository_full_name ='${param.github_repo_full_name}' limit 10"
  }

  output "val" {
    value = step.query.list_issues.rows
  }
}
