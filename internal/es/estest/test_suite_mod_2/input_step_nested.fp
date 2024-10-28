pipeline "input_step_parent" {

    step "input" "my_step" {
        type   = "button"
        prompt = "input_step_parent - Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "pipeline" "nested" {
        pipeline = pipeline.input_step_child
    }

    output "val" {
        value = step.pipeline.nested
    }
}

pipeline "input_step_child" {

    step "sleep" "sleep" {
        duration = "15s"
    }

    step "input" "my_step" {
        depends_on = [step.sleep.sleep]

        type   = "button"
        prompt = "input_step_child - Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }
}

pipeline "parent_with_no_input_step" {

    step "pipeline" "nested" {
        pipeline = pipeline.input_step_child_with_no_sleep
    }

    step "pipeline" "nested_2" {
        pipeline = pipeline.input_step_child_with_no_sleep
    }


    output "val" {
        value = step.pipeline.nested
    }

    output "val_2" {
        value = step.pipeline.nested_2
    }
}

pipeline "input_step_child_with_no_sleep" {

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }
}