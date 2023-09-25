pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    param "param_one" {
        type = string
        default = "this is the default value"
    }

    step "echo" "simple_echo" {
        text = "foo bar: ${param.param_one}"
    }
}

trigger "schedule" "my_every_minute_trigger" {
    schedule = "* * * * *"
    pipeline = pipeline.simple_with_trigger
    args = {
        param_one = "from trigger"
    }
}

trigger "schedule" "my_every_minute_trigger_nine" {
    schedule = "* * * * *"
    pipeline = pipeline.simple_with_trigger
    args = {
        param_one = "from trigger"
    }
}


pipeline "http_webhook_pipeline" {
    param "event" {
        type = string
    }

    step "echo" "simple_echo" {
        text = "event: ${param.event}"
    }

    output "response" {
        value = step.echo.simple_echo
    }
}

trigger "http" "http_trigger" {

    response_body = "ok"

    response_headers = {
      Content-Type = "application/json"
      User-Agent  = "flowpipe"
    }

    pipeline = pipeline.http_webhook_pipeline

    args = {
        event = self.request_body
    }
}