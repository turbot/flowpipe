pipeline "simple_text" {
    description = "text pipeline - debug should be removed"
    step "echo" "text_1" {
        text = "foo"
    }
}

trigger "schedule" "my_every_30_minute_trigger" {
    description = "trigger that will run every 30 minutes"
    schedule    = "*/30 * * * *"
    pipeline    = pipeline.simple_text
}
