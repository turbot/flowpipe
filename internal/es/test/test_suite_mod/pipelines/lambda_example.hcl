pipeline "lambda_example" {

    step "function" "validate_policy_step" {
        function = function.validate_policy
    }

    output "val" {
        value = jsondecode(step.function.validate_policy_step.result)
    }

}
