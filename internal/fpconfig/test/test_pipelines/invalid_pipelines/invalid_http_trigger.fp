trigger "http" "http_trigger_header_example" {

  if = 0 > 1
  pipeline = pipeline.http_webhook_pipeline
  args = {
      event = "foo"
  }

}

pipeline "http_webhook_pipeline" {
    param "event" {
        type = string
    }
    step "echo" "simple_echo" {
        text = "event is: ${param.event}"
    }
}