pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "schedule" "invalid_attribute_test" {
  schedule       = "5 * * * *"
  pipeline       = pipeline.simple_with_trigger
  execution_mode = "synchronous"
}
