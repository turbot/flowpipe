pipeline "simple_with_trigger" {
    step "echo" "simple_echo" {
        text = "foo bar"
    }
}


trigger "interval" "my_hourly_trigger" {
    interval = "hourly"
    pipeline = pipeline.simple
}