trigger "http" "http_trigger_to_iam_policy_validation" {
    pipeline = pipeline.lambda_example
    args     = {
      body   = self.request_body
      headers = self.request_headers
      event = jsondecode(jsondecode(self.request_body).Message)
    }
}

pipeline "lambda_example" {

    # Parameters
    param "aws_region" {
      description = "AWS Region"
      type = string
      default = var.aws_region
    }
    param "aws_access_key_id" {
      description = "AWS Access Key ID"
      type = string
      default = var.aws_access_key_id
    }
    param "aws_secret_access_key" {
      description = "AWS Secret Access Key"
      type = string
      default = var.aws_secret_access_key
    }

    param "body" {
      description = "The body of the request"
      type = string
    }
    param "headers" {
      description = "The headers of the request"
      type = map
    }

    param "restricted_actions" {
      description = "The actions that are not allowed"
      type = string
      default = "s3:DeleteBucket,s3:DeleteObject"
    }

    param "event" {
      description = "The IAM Policy event"
      type = any
    }

    # Steps
    step "http" "confirm_reply" {
      description = "Confirms the SNS topic subscription"
      if = param.headers["X-Amz-Sns-Message-Type"] == "SubscriptionConfirmation"
      method = "get"
      url    = jsondecode(param.body)["SubscribeURL"]
    }

    step "function" "transform_input_step" {
      description = "Transforms the input to the format expected by the validation function"
      runtime = "nodejs:18"
      handler = "index.handler"
      src = "./functions/transform-input"
      event = param.event
    }

    step "function" "validate_policy_step" {
      description = "Validates the IAM Policy event"
      runtime = "nodejs:18"
      handler = "index.handler"
      src = "./functions/validate-policy"
      event = step.function.transform_input_step.result

      env = {
        "restrictedActions" = param.restricted_actions
      }
    }

    step "function" "revert_policy_step" {
      description = "Reverts the IAM Policy event if needed"

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

    # Outputs
    output "transform_returning_message" {
      value = step.function.transform_input_step.result
    }

    output "validation_returning_message" {
      value = step.function.validate_policy_step.result.message
    }

    output "validation_returning_action" {
      value = step.function.validate_policy_step.result.action
    }

    output "reverting_returning_message" {
      value = step.function.revert_policy_step.result.message
    }

}
