// usage: flowpipe pipeline run repository_get_by_full_name
pipeline "repository_get_by_full_name" {
  description = "Get the details of a given repository by the owner and repository name."

  param "token" {
    type    = string
    default = var.token
  }

  param "repository_owner" {
    type    = string
  }

  param "repository_name" {
    type    = string
  }

  step "http" "repository_get_by_full_name" {
    method = "post"
    url    = "https://api.github.com/graphql"
    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.token}"
    }

    request_body = jsonencode({
      query = <<EOQ
        query {
          repository(owner: "${param.repository_owner}", name: "${param.repository_name}") {
            description
            forkCount
            id
            name
            owner {
              id
            }
            stargazerCount
            url
            visibility
          }
        }
        EOQ
    })
  }

  output "repository_id" {
    value = jsondecode(step.http.repository_get_by_full_name.response_body).data.repository.id
  }
  output "stargazer_count" {
    value = jsondecode(step.http.repository_get_by_full_name.response_body).data.repository.stargazerCount
  } 
  output "response_body" {
    value = step.http.repository_get_by_full_name.response_body
  }
  output "response_headers" {
    value = step.http.repository_get_by_full_name.response_headers
  }
  output "status_code" {
    value = step.http.repository_get_by_full_name.status_code
  }
}
