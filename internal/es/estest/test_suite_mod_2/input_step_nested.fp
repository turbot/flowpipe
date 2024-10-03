pipeline "input_step_parent" {

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "pipeline" "nested" {
        pipeline = pipeline.input_step_child
    }
}

pipeline "input_step_child" {

    step "sleep" "sleep" {
        duration = "15s"
    }

    step "input" "my_step" {
        depends_on = [step.sleep.sleep]

        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }
}