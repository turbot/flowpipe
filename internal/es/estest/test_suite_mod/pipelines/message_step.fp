pipeline "message_step_with_throw" {

    step "message" "message" {
        notifier = notifier.default
        text = "This is a message step"

        throw {
            if = result.text == "This is a message step"
            message = "throw here"
        }
    }
}

pipeline "message_step_bad_slack_call" {

    step "message" "message" {
        notifier  = notifier.bare_minimum_slack_wh
        text = "This is a message step"
    }
}

pipeline "message_step_bad_slack_call_ignored" {

    step "message" "message" {
        notifier  = notifier.bare_minimum_slack_wh
        text = "This is a message step"

        error {
            ignore = true
        }
    }
}