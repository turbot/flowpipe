trigger "http" "http_trigger_to_iam_policy_validation" {
    pipeline = pipeline.lambda_example
    args     = {
      body   = self.request_body
      headers = self.request_headers
      event = jsondecode(jsondecode(self.request_body).Message)
    }
}

variable "aws_region" {
    type = string
    default = "asia-southeast1"
}

variable "aws_access_key_id" {
    type = string
    default = "Enter your AWS access key id"
}

variable "aws_secret_access_key" {
    type = string
    default = "Enter your AWS secret access key"
}

pipeline "lambda_example" {

    param "body" {
      type = string
    }
    param "headers" {
      type = map
    }

    step "http" "confirm_reply" {
      if = param.headers["X-Amz-Sns-Message-Type"] == "SubscriptionConfirmation"

      method = "get"
      url    = jsondecode(param.body)["SubscribeURL"]
    }

    param "restricted_actions" {
        type = string
        default = "s3:DeleteBucket,s3:DeleteObject"
    }

    param "aws_region" {
        type = string
        default = var.aws_region
    }
    param "aws_access_key_id" {
        type = string
        default = var.aws_access_key_id
    }
    param "aws_secret_access_key" {
        type = string
        default = var.aws_secret_access_key
    }

    param "event" {
        type = any
    }

    step "function" "transform_input_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/transform-input"
        event = param.event
    }

    output "transform_returning_message" {
        value = step.function.transform_input_step.result
    }

    step "function" "validate_policy_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/validate-policy"
        event = step.function.transform_input_step.result

        env = {
            "restrictedActions" = param.restricted_actions
        }
    }

    output "validation_returning_message" {
        value = step.function.validate_policy_step.result.message
    }

    output "validation_returning_action" {
        value = step.function.validate_policy_step.result.action
    }

    step "function" "revert_policy_step" {
        if = step.function.validate_policy_step.result.action == "remedy"
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/revert-policy"
        event = step.function.transform_input_step.result

        env = {
            "restrictedActions" = param.restricted_actions
            AWS_REGION = param.aws_region
            AWS_ACCESS_KEY_ID = param.aws_access_key_id
            AWS_SECRET_ACCESS_KEY = param.aws_secret_access_key
        }
    }

    output "reverting_returning_message" {
        value = step.function.revert_policy_step.result.message
    }

}
