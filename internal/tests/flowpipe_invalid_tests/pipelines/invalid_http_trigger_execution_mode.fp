trigger "http" "http_trigger_header_example" {

  method "get" {
    pipeline = pipeline.http_webhook_pipeline
    args = {
      event = "foo"
    }

    execution_mode = "foo"
  }
}

pipeline "http_webhook_pipeline" {
  param "event" {
    type = string
  }
  step "transform" "simple_echo" {
    value = "event is: ${param.event}"
  }
}
