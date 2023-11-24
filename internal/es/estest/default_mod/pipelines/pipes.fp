pipeline "pipes_echo" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo" {
        value = step.transform.foo.value
    }
}


pipeline "two_steps" {
    step "transform" "one" {
        value = "Step One"
    }

    step "transform" "two" {
        value = "${step.transform.one.value} => Step Two"
    }

    output "out" {
        value = step.transform.two.value
    }
}

pipeline "three_steps_with_wait" {
    step "transform" "one" {
        value = "Step One"
    }

    step "sleep" "sleep" {
        duration = "1s"
    }

    step "transform" "three" {
        depends_on = [step.sleep.sleep]
        value      = "End"
    }
}


pipeline "use_child_pipeline" {

    step "pipeline" "from_child" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }
}
