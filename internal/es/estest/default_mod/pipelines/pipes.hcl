pipeline "pipes_echo" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo" {
        value = step.echo.foo.text
    }
}


pipeline "two_steps" {
    step "echo" "one" {
        text = "Step One"
    }

    step "echo" "two" {
        text = "${step.echo.one.text} => Step Two"
    }

    output "out" {
        value = step.echo.two.text
    }
}

pipeline "three_steps_with_wait" {
    step "echo" "one" {
        text = "Step One"
    }

    step "sleep" "sleep" {
        duration = "1s"
    }

    step "echo" "three" {
        depends_on = [step.sleep.sleep]
        text = "End"
    }
}


pipeline "use_child_pipeline" {

    step "pipeline" "from_child" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }
}
