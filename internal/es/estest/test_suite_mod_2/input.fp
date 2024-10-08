pipeline "my_step" {

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "transform" "do_the_thing" {
        depends_on = [step.input.my_step]
        value = step.input.my_step.value
    }

    output "val" {
        value = step.transform.do_the_thing
    }
}

pipeline "my_step_2" {

    step "sleep" "sleep" {
        duration = "3s"
    }

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "input" "my_step_2" {
        depends_on = [step.input.my_step]

        type   = "button"
        prompt = "Do you want to approve (2)?"

        option "Approve (2)" {}
        option "Deny (2)" {}

        notifier = notifier.admin
    }



    output "val" {
        value = step.input.my_step_2
    }
}

pipeline "my_step_single" {

    step "input" "my_step" {
        type   = "button"
        prompt = "Do you want to approve?"

        option "Approve" {}
        option "Deny" {}

        notifier = notifier.admin
    }

    step "transform" "do_the_thing" {
        depends_on = [step.input.my_step]
        value = step.input.my_step.value
    }

    output "val" {
        value = step.transform.do_the_thing
    }
}
