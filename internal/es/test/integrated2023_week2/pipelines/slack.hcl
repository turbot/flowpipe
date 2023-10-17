trigger "http" "webhook_trigger" {
    title    = "Webhook Trigger for Slack command"
    pipeline = pipeline.slack_hello

    args     = {
      response_url = parse_query_string(self.request_body).response_url
      command      = parse_query_string(self.request_body).command
      prompt       = parse_query_string(self.request_body).text
      request_body = self.request_body
    }
}


pipeline "slack_hello" {

    param "response_url" {
        description = "The url to respond to slack" 
        type        = string
    }
    param "prompt" {
        description = "The prompt that the user passed to the slack command" 
        type        = string
    }
    param "request_body" {
        description = "The request body that the user passed to the slack command" 
        type        = string
    } 
    param "command" {
        description = "The command that the user used" 
        type        = string
    }   
 
    step "http" "hi_turbie" {
        if = param.prompt == ""
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Hi, I am Turbie, I can create issues for you, just type /hiturbie create issue in <repo_name> <issue title>, and I will return you the issue number."
        })
    }

    # step "http" "check_prompt_value" {
    #     if = startswith(param.prompt, "create issue")
    #     url         = param.response_url
    #     method      = "post"

    #     request_body = jsonencode({
    #         text = "${join(" ", slice(split(" ", param.prompt), 4, length(split(" ", param.prompt))))}"
    #         # text = "${slice(split(" ", param.prompt), 3, length(split(" ", param.prompt)))}"
    #     })
    # }

    # step "pipeline" "hello_world_2" {
    #     if = param.prompt == "blah2"
    #     pipeline = pipeline.slack_hello_2
    #     args = {
    #         response_url = param.response_url
    #     }
    # }

    step "pipeline" "create_issue" {
        if = startswith(param.prompt, "create issue")
        pipeline = pipeline.issue_create
        args = {
            response_url = param.response_url
            repository_name = split(" ",param.prompt)[3]
            issue_title = join(" ", slice(split(" ", param.prompt), 4, length(split(" ", param.prompt))))
        }
    }

    output "create_issue_one" {
        value = step.pipeline.create_issue.repository_get_by_full_name
    }

    # output "request_body" {
    #     value = param.request_body
    # }
}

# pipeline "slack_hello_2" {

#     param "response_url" {
#         description = "The url to respond to slack" 
#         type        = string
#     }

#     step "http" "hello_world_2" {
#         description = "test me"
#         url         = param.response_url
#         method      = "post"

#         request_body = jsonencode({
#             text = "Response from blah step 22222"
#             # text = "bye"
#         })
#     }
# }