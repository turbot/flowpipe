
integration "slack" "my_slack_app" {
  token           = var.slack_token
}


pipeline "user_group_add" {

  param "username" {
    type        = string 
    description = "The username of the user to add to the group"
  }

  param "group" {
    type        = string 
    description = "The group that the user should be added to"
  }

  step "container" "before" {
    image = "amazon/aws-cli"
    cmd   = ["iam", "list-groups-for-user", "--user-name", param.username]
    env   = local.aws_creds_vars
  }


  step "input" "approval" {
    type       = "slack"
    channel    = "#mantix_test"
    slack_type = "button"
    prompt     = "User ${param.username} has requested to be added to the ${param.group} group. Do you want to approve?"
    options    = ["Approve","Deny"]

    notify {
      integration = integration.slack.my_slack_app
    #  channel    = "#mantix_test"
    }
  }

  step "container" "add_user_to_group" {
    if          = step.input.approval.value == "Approve"
    depends_on  = [step.container.before]
    description = "Run the AWS cli command to add the user"
    image       = "amazon/aws-cli"
    cmd         = ["iam", "add-user-to-group", "--user-name", param.username, "--group-name", param.group]
    env         = local.aws_creds_vars
  }

  step "container" "after" {
    depends_on = [step.container.add_user_to_group]
    image      = "amazon/aws-cli"
    cmd        = ["iam", "list-groups-for-user", "--user-name", param.username]
    env        = local.aws_creds_vars
  }

  output "cli_stdout" {
    value = step.container.add_user_to_group.stdout
  }

  output "cli_stderr" {
    value = step.container.add_user_to_group.stdout
  }

  output "before" {
    value = jsondecode(step.container.before.stdout)
  }
  output "after" {
    value = jsondecode(step.container.after.stdout)
  }

  output "approved" {
    value = step.input.approval.value == "Approve"
  }

}


