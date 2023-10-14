pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

pipeline "simple_with_trigger_two" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

trigger "schedule" "my_hourly_trigger" {
    schedule = "5 * * * *"
    pipeline = pipeline.simple_with_trigger
}

trigger "schedule" "my_hourly_trigger" {
    schedule = "5 * * * *"
    pipeline = pipeline.simple_with_trigger_two
}
