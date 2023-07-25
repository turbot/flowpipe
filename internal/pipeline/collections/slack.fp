
pipeline "send_slack_message" {
    description = "johns pipeline"

    param "webhook_url" {
        description = "Slack webhook URL"
        default = "https://hooks.slack.com/services/<change me>"
    }
    param "message" {
        description = "message to send"
        default = "this is a test"
    }

    step "http" "send_message" {
        url    = param.webhook_url
        method = "POST"
        request_body   = "{ \"text\": \"${param.message}\"}"
    }

    output "response" {
        value = step.send_message.body_json
    }
}
