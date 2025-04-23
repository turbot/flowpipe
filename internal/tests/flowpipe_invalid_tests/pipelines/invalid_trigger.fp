
trigger "schedule" "trigger_bad_ref" {
    schedule = "5 * * * * *"
    pipeline = pipeline.bad_pipeline
}

trigger "schedule" "trigger_no_pipelines" {
    schedule = "5 * * * * *"
}

pipeline "simple_with_trigger_a" {
    description = "simple pipeline that will be referred to by a trigger"

    step "transform" "simple_echo" {
        value = "foo bar"
    }
}

