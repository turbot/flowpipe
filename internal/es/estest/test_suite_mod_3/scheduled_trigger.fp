
trigger "schedule" "s_simple" {
    schedule = "1 * * * *"
    pipeline = pipeline.simple
}

pipeline "simple" {

    step "transform" "echo" {
        value = "hello world"
    }

    output "val" {
        value = step.transform.echo
    }
}


trigger "schedule" "s_simple_failure" {
    schedule = "1 * * * *"
    pipeline = pipeline.simple_failure
}

pipeline "simple_failure" {

    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"
    }

    output "val" {
        value = "should not be calculated"
    }
}

trigger "schedule" "s_simple_error_ignored_with_if_matches" {
    schedule = "1 * * * *"
    pipeline = pipeline.simple_error_ignored_with_if_matches
}

pipeline "simple_error_ignored_with_if_matches" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"

        error {
            if = result.status_code == 404
            ignore = true
        }
    }

    output "val" {
        value = "should be calculated"
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
