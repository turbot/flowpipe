
// integration "slack" "my_slack_app_hanging" {
//   token = "abcdefg"
// }

// pipeline "test_integration_hanging" {
//   step "transform" "initial_step" {
//     value = "Awaiting Input..."
//   }

//   step "input" "get_value" {

//     prompt = "Pick One"
//     options = ["Hello","Goodbye","What?"]
//     notify {
//       integration = integration.slack.my_slack_app_hanging
//       channel     = "#general"
//     }
//   }
// }