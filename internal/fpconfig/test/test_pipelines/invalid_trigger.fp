
trigger "schedule" "trigger_bad_ref" {
    schedule = "5 * * * * *"
    pipeline = pipeline.bad_pipeline
}

trigger "schedule" "trigger_no_pipelines" {
    schedule = "5 * * * * *"
}

pipeline "simple_with_trigger_a" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

trigger "schedule" "trigger_bad_cron" {
    schedule = "bad cron format"
    pipeline = pipeline.simple_with_trigger_a
}

trigger "interval" "trigger_bad_interval" {
    schedule = "bad interval format"
    pipeline = pipeline.simple_with_trigger_a
}
