// integration "slack" "my_slack" {
//   token = "xoxp-fake-token"
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