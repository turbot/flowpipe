pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "http" "http_trigger_with_self" {
  method "post" {
    pipeline = pipeline.simple_with_trigger

    args = {
      event = self.request_body
    }
  }
}


