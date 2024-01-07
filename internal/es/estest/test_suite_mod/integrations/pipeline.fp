// pipeline "my_pipe" {

//   step "input" "my_step" {

//     notify  {
//       integration = integration.slack.my_slack_app
//       channel     = "#bluth-banana"
//     }

//     type   = "button"
//     prompt = "do you want to approve?"

//     option "Approve" {}
//     option "Deny" {}

//   }

//   step "pipeline" "do_the_thing" {
//     pipeline = pipeline.something
//     if       = step.input.my_step.value == "Approve"
//   }
// }