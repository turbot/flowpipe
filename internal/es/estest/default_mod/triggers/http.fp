trigger "http" "my_webhook" {
  pipeline = pipeline.my_pipeline
  execution_mode = "synchronous"
  args     = {
    event = self.request_body
  }
}