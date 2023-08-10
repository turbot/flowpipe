pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

trigger "schedule" "my_every_minute_trigger" {
    schedule = "* * * * *"
    pipeline = pipeline.simple_with_trigger
}

trigger "schedule" "my_every_5_minute_trigger" {
    schedule = "*/5 * * * *"
    pipeline = pipeline.simple_with_trigger
}

trigger "interval" "every_hour_trigger" {
    schedule = "hourly"
    pipeline = pipeline.simple_with_trigger
}