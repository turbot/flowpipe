
trigger "schedule" "s_simple" {
    schedule = "1 * * * *"
    pipeline = pipeline.my_step
}


pipeline "simple" {

    step "transform" "echo" {
        value = "hello world"
    }

    output "val" {
        value = step.transform.echo
    }
}


trigger "schedule" "my_step" {
    schedule = "1 * * * *"
    pipeline = pipeline.my_step
}


pipeline "my_step" {

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

    step "transform" "do_the_thing" {
        depends_on = [step.input.my_step]
        value = step.input.my_step.value
    }

    output "val" {
        value = step.transform.do_the_thing
    }
}
