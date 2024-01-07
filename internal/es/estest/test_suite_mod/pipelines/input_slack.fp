// integration "slack" "my_slack" {
//   token = "xoxp-hkfkhfkafha131424255"
// }

// pipeline "input_slack_pipeline" {
//   step "input" "my_step" {
//     # slack_type = "button"
//     prompt = "Do you want to approve?"

//     // option "Approve" {}
//     // option "Deny" {}

//     notify {
//       integration = integration.slack.my_slack
//       channel     = "DF8SL4GR5"
//     }
//   }
// }