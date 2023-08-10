pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

trigger "schedule" "my_hourly_trigger" {
    schedule = "5 * * * * *"
    pipeline = pipeline.simple_with_trigger
}


trigger "schedule" "trigger_with_args" {
    schedule = "5 * * * * *"
    pipeline = pipeline.simple_with_trigger

    args = {
        param_one     = "one"
        param_two_int = 2
    }
}