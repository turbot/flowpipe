pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "schedule" "my_hourly_trigger" {
  schedule = var.trigger_schedule
  pipeline = pipeline.simple_with_trigger
}

