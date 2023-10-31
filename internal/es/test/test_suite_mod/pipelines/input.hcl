integration "slack" "dev_app" {
    token = "abcde"
}

integration "email" "dev_workspace" {
    smtp_host       = "foo bar baz"
    default_subject = "bar foo baz"
    smtp_username   = "baz bar foo"
}

pipeline "input_one" {
    step "echo" "echo" {
        text = "hello"
    }

    step "input" "input" {
        // type = button
        // destination = slack

    }
}

pipeline "input_notify" {
    param "channel" {
        type = string
        default = "#general"
    }

    step "input" "input" {

        prompt = "Choose an option:"

        notify {
            integration = integration.slack.dev_app
            channel     = param.channel
        }
    }
}

pipeline "input_notifies" {
    param "channel" {
        type = string
        default = "#general"
    }

    step "input" "input" {

        prompt = "Choose an option:"

        notifies = [
            {
                integration = integration.slack.dev_app
                channel     = param.channel
            },
            {
                integration = integration.email.dev_workspace
                to          = "awesomebob@blahblah.com"
            }
        ]
    }
}