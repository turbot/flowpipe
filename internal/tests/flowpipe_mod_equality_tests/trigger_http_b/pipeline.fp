trigger "http" "my_webhook" {

  method "post" {
    pipeline = pipeline.my_pipeline

    args = {
      event = self.request_body
    }
  }
    
  method "get" {
    pipeline       = pipeline.confirm_setup

    args = {
      headers = self.request_headers
    }
  }
                              
}

pipeline "my_pipeline" {
  param "event" {
  }

  step "transform" "echo" {
    value = param.event
  }

  output "val" {
    value = step.transform.echo.value
  }
}

pipeline "confirm_setup" {
  param "headers" {
  }

  step "transform" "echo" {
    value = param.headers
  }

  output "val" {
    value = step.transform.echo.value
  }
}
