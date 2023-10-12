pipeline "lambda_example" {

    param "event" {
        type = any
    }

    step "function" "validate_policy_step" {
        runtime = "nodejs:18"
        handler = "index.handler"
        src = "./functions/validate-policy"
        event = param.event
    }

    output "returning_message" {
        value = step.function.validate_policy_step.result.message
    }

}
