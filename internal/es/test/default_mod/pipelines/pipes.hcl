pipeline "pipes_echo" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo" {
        value = step.echo.foo.text
    }
}


pipeline "use_child_pipeline" {

    step "pipeline" "from_child" {
        pipeline = mod_depend_a.pipeline.echo_one_depend_a
    }
}

trigger "schedule" "my_every_minute_trigger" {
    schedule = "* * * * *"
    pipeline = pipeline.use_child_pipeline
    args = {
        param_one = "from trigger"
    }
}
