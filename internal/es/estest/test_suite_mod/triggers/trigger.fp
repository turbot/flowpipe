pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    param "param_one" {
        type = string
        default = "this is the default value"
    }

    step "transform" "simple_echo" {
        value = "foo bar: ${param.param_one}"
    }
}

trigger "schedule" "my_every_hour_trigger" {
    schedule = "1h"
    pipeline = pipeline.simple_with_trigger
    args = {
        param_one = "from trigger"
    }
}

// trigger "schedule" "my_every_minute_trigger_nine" {
//     schedule = "* * * * *"
//     pipeline = pipeline.simple_with_trigger
//     args = {
//         param_one = "from trigger"
//     }
// }

// trigger "schedule" "my_every_minute_with_var" {
//     schedule = "* * * * *"
//     pipeline = pipeline.simple_with_trigger
//     args = {
//         param_one = var.var_two
//     }
// }


// pipeline "http_webhook_pipeline" {
//     param "event" {
//         type = string
//     }

//     step "transform" "simple_echo" {
//         value = "event: ${param.event}"
//     }

//     output "output_from_request_body" {
//         value = step.transform.simple_echo
//     }

//     output "hardcoded_output" {
//         value = "foo"
//     }
// }

// trigger "http" "http_trigger" {

//     pipeline = pipeline.http_webhook_pipeline

//     args = {
//         event = self.request_body
//     }
// }

// trigger "http" "http_trigger_header" {

//     pipeline = pipeline.http_webhook_pipeline

//     args = {
//         event = self.request_headers["X-Event"]
//     }
// }