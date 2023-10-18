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
            text = "Hi, I am Turbie, I can create issues for you, just type /hiturbie create issue in <repo_name> <issue title>, I can create a new TE stack, please type /hiturbie create TE version <version: ex 5.42.3> in <region: ex ap-southeast-1>. For TEF update, please type /hiturbie update TEF to version <version: ex 5.42.3> in <region: ex ap-southeast-1>. For TED update, please type /hiturbie update TED <hive> to version <version: ex 5.42.3> in <region: ex ap-southeast-1>. For TED creation, please type /hiturbie create TED <hive> with version <version: ex 1.37.0> in <region: ex ap-southeast-1>"
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

    step "pipeline" "trigger_te_creation" {
        if = startswith(param.prompt, "create TE version")
        pipeline = pipeline.create_te
        args = {
            response_url = param.response_url
            version = split(" ",param.prompt)[3]
            region = split(" ",param.prompt)[5]
        }
    }

    step "pipeline" "trigger_tef_update" {
        if = startswith(param.prompt, "update TEF to version")
        pipeline = pipeline.update_tef
        args = {
            response_url = param.response_url
            version = split(" ",param.prompt)[4]
            region = split(" ",param.prompt)[6]
        }
    }

    step "pipeline" "trigger_ted_update" {
        if = startswith(param.prompt, "update TED")
        pipeline = pipeline.update_ted
        args = {
            response_url = param.response_url
            hive = split(" ",param.prompt)[2]
            version = split(" ",param.prompt)[4]
            region = split(" ",param.prompt)[6]
        }
    }

    step "pipeline" "trigger_ted_creation" {
        if = startswith(param.prompt, "create TED")
        pipeline = pipeline.create_ted
        args = {
            response_url = param.response_url
            hive = split(" ",param.prompt)[2]
            version = split(" ",param.prompt)[4]
            region = split(" ",param.prompt)[6]
        }
    }

    # step "pipeline" "create_issue" {
    #     if = startswith(param.prompt, "create issue")
    #     pipeline = pipeline.issue_create
    #     args = {
    #         response_url = param.response_url
    #         repository_name = split(" ",param.prompt)[3]
    #         issue_title = join(" ", slice(split(" ", param.prompt), 4, length(split(" ", param.prompt))))
    #     }
    # }

    # output "create_issue_one" {
    #     value = step.pipeline.create_issue.repository_get_by_full_name
    # }

    # output "request_body" {
    #     value = param.request_body
    # }
}

pipeline "create_te" {

    param "response_url" {
        description = "The url to respond to slack"
        type        = string
    }

    param "version" {
        description = "The TE stack version" 
        type        = string
    }

    param "region" {
        description = "The TE stack version" 
        type        = string
    }

    step "http" "on_it" {
        description = "Echo the user's request back to them"
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Creating TE ${param.version}..."
        })
    }

    step "container" "create_te" {
        description = "Run the AWS CLI command in the aws-cli container"
        image       = "aws-cli-image-pritha-v1"
        cmd         = ["aws", "servicecatalog", "provision-product", "--product-name", "Turbot Guardrails Enterprise", "--provisioning-artifact-name", "v${param.version}", "--provisioning-parameters", "file://root/parameter.json", "--provisioned-product-name", "te-${split(".",param.version)[0]}-${split(".",param.version)[1]}-${split(".",param.version)[2]}", "--region", "${param.region}"]
    }

    step "http" "done" {
        description = "Respond to slack with the acknowledgement"
        depends_on  = [step.container.create_te]
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "TE creation started"
        })
    }
}

pipeline "update_tef" {

    param "response_url" {
        description = "The url to respond to slack"
        type        = string
    }

    param "version" {
        description = "The TEF stack version" 
        type        = string
    }

    param "region" {
        description = "The TEF stack version" 
        type        = string
    }

    step "http" "on_it" {
        description = "Echo the user's request back to them"
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Updating TEF to ${param.version}..."
        })
    }

    step "container" "update_tef" {
        description = "Run the AWS CLI command in the aws-cli container"
        image       = "aws-cli-image-pritha-v1"
        cmd         = ["aws", "servicecatalog", "update-provisioned-product", "--product-name", "Turbot Guardrails Enterprise Foundation", "--provisioned-product-name", "tef-helix", "--provisioning-artifact-name", "v${param.version}", "--provisioning-parameters", "file://root/parameter_tef.json", "--region", "${param.region}"]
    }

    step "http" "done" {
        description = "Respond to slack with the acknowledgement"
        depends_on  = [step.container.update_tef]
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "TEF update started"
        })
    }
}

pipeline "update_ted" {

    param "response_url" {
        description = "The url to respond to slack"
        type        = string
    }

    param "version" {
        description = "The TED stack version" 
        type        = string
    }

    param "region" {
        description = "Region" 
        type        = string
    }

    param "hive" {
        description = "Hive name" 
        type        = string
    }

    step "http" "on_it" {
        description = "Echo the user's request back to them"
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Updating TED ${param.hive} to ${param.version}..."
        })
    }

    step "container" "update_ted" {
        description = "Run the AWS CLI command in the aws-cli container"
        image       = "aws-cli-image-pritha-v1"
        // cmd         = ["aws", "s3", "ls"]
        // cmd = []
        cmd         = ["aws", "servicecatalog", "update-provisioned-product", "--product-name", "Turbot Guardrails Enterprise Database", "--provisioned-product-name", "ted", "--provisioning-artifact-name", "v${param.version}", "--provisioning-parameters", "file://root/parameter_ted.json", "--region", "${param.region}"]
    }

    step "http" "done" {
        description = "Respond to slack with the acknowledgement"
        depends_on  = [step.container.update_ted]
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "TED update started"
        })
    }
}

pipeline "create_ted" {

    param "response_url" {
        description = "The url to respond to slack"
        type        = string
    }

    param "version" {
        description = "The TE stack version" 
        type        = string
    }

    param "region" {
        description = "The TE stack version" 
        type        = string
    }

    param "hive" {
        description = "Hive name" 
        type        = string
    }

    step "http" "on_it" {
        description = "Echo the user's request back to them"
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "Creating TED ${param.version}..."
        })
    }

    step "container" "create_ted" {
        description = "Run the AWS CLI command in the aws-cli container"
        image       = "aws-cli-image-pritha-v1"
        cmd         = ["aws", "servicecatalog", "provision-product", "--product-name", "Turbot Guardrails Enterprise Database", "--provisioning-artifact-name", "v${param.version}", "--provisioning-parameters", "file://root/parameter_ted_create.json", "--provisioned-product-name", "ted-${param.hive}", "--region", "${param.region}"]
    }

    step "http" "done" {
        description = "Respond to slack with the acknowledgement"
        depends_on  = [step.container.create_ted]
        url         = param.response_url
        method      = "post"

        request_body = jsonencode({
            text = "TED creation started"
        })
    }
}

