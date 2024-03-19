pipeline "input_step_error_out" {
    step "input" "test" {
        notifier = notifier.bad_slack_notifier
        type     = "button"
        prompt   = "do you approve?"

        option "Approve" {}

        option "Deny" {}
    }
}

pipeline "input_step_error_out_error_config" {
    step "input" "test" {
        notifier = notifier.bad_slack_notifier
        type     = "button"
        prompt   = "do you approve?"

        option "Approve" {}

        option "Deny" {}

        error {
            ignore = true
        }
    }
}

pipeline "input_step_error_out_retry" {
    step "input" "test" {
        notifier = notifier.bad_slack_notifier
        type     = "button"
        prompt   = "do you approve?"

        option "Approve" {}

        option "Deny" {}

        retry {
        max_attempts = 3
        }
    }
}