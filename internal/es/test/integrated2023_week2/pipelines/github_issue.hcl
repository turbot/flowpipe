// usage: flowpipe pipeline run issue_create --pipeline-arg "issue_title=[SUPPORT] please help" --pipeline-arg "issue_body=I need help with..."
pipeline "issue_create" {
    description = "Create a new issue."

    param "token" {
        type    = string
        default = var.token
    }

    param "response_url" {
        type    = string
    }

    # param "repository_owner" {
    #     type    = string
    #     default = "turbotio"
    # }

    param "repository_name" {
        type    = string
    }

    param "issue_title" {
        type = string
    }

    # param "issue_body" {
    #     type = string
    # }

    step "pipeline" "repository_get_by_full_name" {
        pipeline = pipeline.repository_get_by_full_name
        args = {
            token = var.token
            repository_owner = split("/",param.repository_name)[0]
            repository_name  = split("/",param.repository_name)[1]
        }
    }

    output "repository_get_by_full_name" {
        value = step.pipeline.repository_get_by_full_name
    }

    step "http" "create_issue_in_github_step" {
        method = "post"
        url    = "https://api.github.com/graphql"

        request_headers = {
            Content-Type  = "application/json"
            Authorization = "Bearer ${param.token}"
        }

        request_body = jsonencode({
          query = <<EOQ
            mutation {
                createIssue(input: {repositoryId: "${step.pipeline.repository_get_by_full_name.repository_id}", title: "${param.issue_title}"}
                ) {
                    clientMutationId
                    issue {
                      id
                      url
                    }
                }    
            }
            EOQ
        })
    }

    # step "http" "response_to_slack" {
    #     description = "test me"
    #     url         = param.response_url
    #     method      = "post"

    #     request_body = jsonencode({
    #         # text = "${step.http.issue_create.response_body}"
    #         # text = "issue created in github: ${jsondecode(${step.http.issue_create.response_body}).data.createIssue.issue.id} "
    #         text = "issue created"
    #     })
    # }

    # output "repository_id" {
    #     value = jsondecode(step.http.repository_get_by_full_name.response_body).data.repository.id
    # }
#   output "issue_id" {
#     value = jsondecode(step.http.issue_create.response_body).data.createIssue.issue.id
#   }
#   output "response_body" {
#     value = step.http.issue_create.response_body
#   }
#   output "response_headers" {
#     value = step.http.issue_create.response_headers
#   }
#   output "status_code" {
#     value = step.http.issue_create.status_code
#   }

}
