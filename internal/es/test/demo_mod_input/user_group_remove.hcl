
pipeline "user_group_remove" {

  param "username" {
    type        = string 
    description = "The username of the user to remove from the group"
  }

  param "group" {
    type        = string 
    description = "The group that the user should be removed from"
  }

  step "container" "before" {
    image = "amazon/aws-cli"
    cmd  = ["iam", "list-groups-for-user", "--user-name", param.username]
    env  = local.aws_creds_vars
  }

  step "input" "approval" {
    type         = "email"
    to           = ["john@turbot.com"]
    subject      = "Approval Requested"
    prompt       = "User ${param.username} has requested to be removed from the ${param.group} group. Do you want to approve?"
    options      = ["Approve","Deny"]

    username     = var.smtp_username
    password     = var.smtp_password
    smtp_server  = var.smtp_server
    smtp_port    = var.smtp_port
    sender_name  = var.smtp_from
    response_url = var.response_url
    
  }

  // step "input" "approval" {
  //   type       = "slack"
  //   token      = var.slack_token
  //   channel    = "#mantix_test"  #"DFL73HHHB"    #"DF8SL4GR5"
  //   slack_type = "button"
  //   prompt     = "User ${param.username} has requested to be removed from the ${param.group} group. Do you want to approve?"
  //   options    = ["Approve","Deny"]
  // }

  step "container" "remove_user_from_group" {
    if          = step.input.approval.value == "Approve"
    depends_on  = [step.container.before]
    description = "Run the AWS cli command to remove the user"
    image       = "amazon/aws-cli"
    cmd         = ["iam", "remove-user-from-group", "--user-name", param.username, "--group-name", param.group]
    env  = local.aws_creds_vars
  }

  step "container" "after" {
    depends_on = [step.container.remove_user_from_group]
    image       = "amazon/aws-cli"
    cmd         = ["iam", "list-groups-for-user", "--user-name", param.username]
    env  = local.aws_creds_vars
  }

  output "cli_stdout" {
    value = step.container.remove_user_from_group.stdout
  }

  output "cli_stderr" {
    value = step.container.remove_user_from_group.stdout
  }

  output "before" {
    value = jsondecode(step.container.before.stdout)
  }
  output "after" {
    value = jsondecode(step.container.after.stdout)
  }
}