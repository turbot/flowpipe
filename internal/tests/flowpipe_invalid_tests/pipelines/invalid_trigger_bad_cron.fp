pipeline "simple_with_trigger_a" {
    description = "simple pipeline that will be referred to by a trigger"

    step "transform" "simple_echo" {
        value = "foo bar"
    }
}

trigger "schedule" "trigger_bad_cron" {
    schedule = "bad cron format"
    pipeline = pipeline.simple_with_trigger_a
}
