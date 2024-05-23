
pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "transform" "simple_echo" {
        value = "foo bar"
    }
    output "echo" {
        value = step.transform.simple_echo.value
    }
}


// integration "slack" "my_slack_app" {
//     webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
//     default_channel     = "#admins"
//     token       = "xoxp-1111111111"
//     url         = ""
// }


trigger "http" "my_app" {
    pipeline = pipeline.simple_with_trigger

    args = {
        param_one     = "one"
        param_two_int = 2
    }
}
