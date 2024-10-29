pipeline "http_webhook_pipeline" {
  param "event" {
    type = string
  }
  step "transform" "simple_echo" {
    value = "event is: ${param.event}"
  }
}

trigger "http" "invalid_http_trigger_method" {

  method "test" {
    pipeline = pipeline.http_webhook_pipeline

    args = {
      event = "test"
    }
  }

}
