pipeline "aws_cli_from_prompt" {
    param "response_url" {
        type = string
    }
    param "prompt" {
        type = string
    }
    step "http" "first_slack_response" {
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = "Your Command: ${param.prompt}"
        })
    }
    step "http" "open_ai" {
        url = "https://api.openai.com/v1/chat/completions"
        method = "post"
        request_headers = {
            Content-Type = "application/json"
            Authorization = "Bearer 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"
        }
        request_body = jsonencode({
            model = "gpt-3.5-turbo"
            temperature = 0.2
            messages = [{
                role = "user"
                content = <<EOQ
                I'd like you to take the command below written in plain text and convert it into an AWS CLI command to perform the requested command. Rules for your response:
                - Return only the AWS CLI command. Do not add text. Do not explain yourself. Do not format.
                - I absolutely strictly need it as a JSON array of strings to be passed to the aws-cli docker container.
                - The AWS CLI command must be syntactically correct and accurate so that it can be run as is.
                - If the request will be destructive (e.g. delete a resource) then add an extra item to the start of the array with the string DRY_RUN_ONLY.
                ${param.prompt}
                EOQ
            }]
        })
    }

    step "http" "done" {
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = jsondecode(step.http.open_ai.response_body).choices[0].message.content
        })
    }

    step "container" "container_run_aws_cli" {
        depends_on = [step.http.open_ai]
        image = "amazon/aws-cli"
        // cmd = ["s3api", "put-bucket-versioning", "--bucket", "testsumitbucket356655", "--versioning-configuration", "Status=Enabled"]
        // cmd = jsonencode(jsondecode(step.http.open_ai.response_body).choices[0].message.content)
        // cmd = jsondecode(jsondecode(step.http.open_ai.response_body).choices[0].message.content)
        cmd = jsondecode(jsondecode(step.http.open_ai.response_body).choices[0].message.content)[0] == "aws" ? slice(jsondecode(jsondecode(step.http.open_ai.response_body).choices[0].message.content), 1, length(jsondecode(jsondecode(step.http.open_ai.response_body).choices[0].message.content)) - 1 ) : jsondecode(jsondecode(step.http.open_ai.response_body).choices[0].message.content)
        // cmd = ["s3api", "list-buckets", "--query", "Buckets[?Versioning.Status!='Enabled'].Name"]
        env = {
            AWS_REGION = "us-east-1"
            AWS_ACCESS_KEY_ID = "AKIAQGDRKHTKBKCJASUB"
            AWS_SECRET_ACCESS_KEY = "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"
        }
    }

    step "http" "done_aws" {
        depends_on = [step.container.container_run_aws_cli]
        url = param.response_url
        method = "post"
        request_body = jsonencode({
            text = length(step.container.container_run_aws_cli.stdout) > 0 ? step.container.container_run_aws_cli.stdout : "${param.prompt} complete! 🚀"
        })
    }

    output "response_url" {
        value = param.response_url
    }
}

trigger "http" "http_trigger_to_open_ai_prompt" {
    pipeline = pipeline.aws_cli_from_prompt
    args     = {
      response_url = parse_query_string(self.request_body).response_url
      prompt = parse_query_string(self.request_body).text
    }
}


