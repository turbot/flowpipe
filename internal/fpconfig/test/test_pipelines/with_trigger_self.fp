pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "echo" "simple_echo" {
        text = "foo bar"
    }
}

trigger "http" "http_trigger_with_self" {
    pipeline = pipeline.simple_with_trigger

    args = {
        event     = self.request_body
    }
}


