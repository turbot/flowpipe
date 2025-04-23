mod "test" {

}

pipeline "input_with_loop" {

    step "input" "my_input" {
        prompt   = "Shall we play a game?"
        type     = "select"
        notifier = notifier.default

        option "Tic Tac Toe" {}
        option "Checkers" {}
        option "Global Thermonuclear War" {}

        loop {
            until = loop.index > 2
            notifier = notifier["notifier_${loop.index}"]
        }
    }
}


pipeline "input_with_loop_2" {

    param "approvers" {

    }
    
    param "subject" {

    }

    param "prompt" {

    }

    step "input" "approve" {
        notifier = notifier[keys(param.approvers)[0]]

        type = "button"

        subject = param.subject

        prompt = param.prompt

        option "approve" {
            label = "Approve"
            style = "ok"
        }

        option "deny" {
            label = "Deny"
            style = "alert"
        }

        loop {
            until = result.value == "deny" || loop.index >= length(param.approvers)
            notifier = notifier[keys(param.approvers)[loop.index]]
        }

    }
}