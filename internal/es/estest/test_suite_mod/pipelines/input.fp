integration "slack" "dev_app" {
    token = "abcde"
}

integration "email" "dev_workspace" {
    smtp_host       = "foo bar baz"
    default_subject = "bar foo baz"
    smtp_username   = "baz bar foo"
}

// pipeline "input_one" {
//     step "transform" "echo" {
//         value = "hello"
//     }

//     step "input" "input" {
//     }
// }

// pipeline "input_slack_notify" {
//     param "channel" {
//         type = string
//         default = "#general"
//     }

//     step "input" "input" {

//         prompt = "Choose an option:"

//         notify {
//             integration = integration.slack.dev_app
//             channel     = param.channel
//         }
//     }
// }

// pipeline "input_email_notify" {
//     param "to" {
//         type = string
//         default = "awesomebob@blahblah.com"
//     }

//     step "input" "input" {

//         prompt = "Choose an option:"

//         notify {
//             integration = integration.email.dev_workspace
//             to          = param.to
//         }
//     }
// }

// pipeline "input_notifies" {
//     param "channel" {
//         type = string
//         default = "#general"
//     }

//     step "input" "input" {

//         prompt = "Choose an option:"

//         notifies = [
//             {
//                 integration = integration.slack.dev_app
//                 channel     = param.channel
//             },
//             {
//                 integration = integration.email.dev_workspace
//                 to          = "awesomebob@blahblah.com"
//             }
//         ]
//     }
// }