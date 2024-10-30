trigger "http" "http_trigger_header_example" {

  if = 0 > 1

  method "post" {
    pipeline = pipeline.http_webhook_pipeline
    args = {
      event = "foo"
    }
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
