
pipeline "send_slack_message" {
    description = "johns pipeline"

    param "webhook_url" {
        description = "Slack webhook URL"
        default = "https://hooks.slack.com/services/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"
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
        value = step.send_message.response_body
    }
}
