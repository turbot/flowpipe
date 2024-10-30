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