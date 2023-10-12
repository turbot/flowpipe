pipeline "lambda_example" {

    param "restricted_actions" {
        type = string
        default = "s3:DeleteBucket,s3:DeleteObject"
    }

    param "event" {
        type = any
    }

    step "function" "validate_policy_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/validate-policy"
        event = param.event

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
        event = param.event

        env = {
            "restrictedActions" = param.restricted_actions
        }
    }

    output "reverting_returning_message" {
        value = step.function.revert_policy_step.result.message
    }

}
