
trigger "http" "priv_cmd" {
  title    = "Webhook Trigger for Slack /priv command"
  pipeline = pipeline.priv_command_router
  args     = {
    body          = self.request_body
  }
}


pipeline "priv_command_router" {

  param "body" {}

  step "echo" "parse" {
    text = param.body

    output "requestor" {
      value = parse_query_string(param.body).user_name
    }

    output "command" {
      value = split("+", parse_query_string(param.body).text)
    }

    output "channel_id" {
      value = parse_query_string(param.body).channel_id
    }

    output "response_url" {
      value = parse_query_string(param.body).response_url
    }
  }

  step "pipeline" "add" {
    if       = length(step.echo.parse.output.command) > 0 && step.echo.parse.output.command[0] == "add"
    pipeline = pipeline.priv_command_add

    args = {
      username   = step.echo.parse.output.requestor
      channel_id = step.echo.parse.output.channel_id
      response_url = step.echo.parse.output.response_url
    }
  }

  step "pipeline" "remove" {
    if       = length(step.echo.parse.output.command) > 0 && step.echo.parse.output.command[0] == "remove"
    pipeline = pipeline.priv_command_remove

    args = {
      username   = step.echo.parse.output.requestor
      channel_id = step.echo.parse.output.channel_id
      response_url = step.echo.parse.output.response_url
    }
  }

  step "pipeline" "list" {
    if       = length(step.echo.parse.output.command) > 0 && step.echo.parse.output.command[0] == "list"
    pipeline = pipeline.priv_command_list

    args = {
      username   = step.echo.parse.output.requestor
      channel_id = step.echo.parse.output.channel_id
      response_url = step.echo.parse.output.response_url
    }
  }

}


####    Add     ####

pipeline "priv_command_add" {

  param "response_url" {}
  param username {}
  param channel_id {}

  step "query" "list_groups" {
    connection_string = "postgres://steampipe@127.0.0.1:9193/steampipe"
    sql = <<-EOQ
      select
        name as label,
        arn as value
      from
        aws_iam_group    
    EOQ
  }

  step "input" "select_group" {
    type       = "slack"
    #token      = var.slack_token
    channel    = param.channel_id # "#mantix_test" #step.echo.parse.output.channel_id
    slack_type = "button"
    prompt     = "Select the group that you would like to join:"
    options    = step.query.list_groups.rows[*].label

    notify {
      integration = integration.slack.my_slack_app
    #  channel    = "#mantix_test"
    }
  }

  step "pipeline" "add_user_to_group" {
    pipeline = pipeline.user_group_add

    args = {
      username = param.username #step.echo.parse.output.requestor
      group    = step.input.select_group.value
    }
  }

  step "http" "respond" {
    description = "Respond to requestor with the results"
    url         = param.response_url
    method      = "post"

    request_body = jsonencode({
      text = "Your request has been ${step.pipeline.add_user_to_group.approved == true ? "approved." : "denied."}"
    })
  }

  output "groups" {
      value = step.query.list_groups.rows
  }

  output "user_group_add" {
    value = step.pipeline.add_user_to_group.output
  }
}



####    Remove     ####

pipeline "priv_command_remove" {

  param "response_url" {}
  param username {}
  param channel_id {}

  step "query" "list_groups" {
    connection_string = "postgres://steampipe@127.0.0.1:9193/steampipe"
    sql = <<-EOQ
      select 
        -- name, 
        g ->> 'GroupName' as label
      from
        aws_iam_user,
        jsonb_array_elements(groups) as g
      where
        lower(name) = lower('${param.username}')
    EOQ
  }

  step "input" "select_group" {
    type       = "slack"
    #token      = var.slack_token
    channel    = param.channel_id
    slack_type = "button"
    prompt     = "You are a member of the following groups.  Select a group to remove:"
    options    = step.query.list_groups.rows[*].label

    notify {
      integration = integration.slack.my_slack_app
    #  channel    = "#mantix_test"
    }
  }

  step "pipeline" "remove_user_from_group" {
    pipeline = pipeline.user_group_remove

    args = {
      username      = param.username #step.echo.parse.output.requestor
      group         = step.input.select_group.value
      auto_approve  = true
    }
  }

  output "groups" {
      value = step.query.list_groups.rows
  }

  output "user_group_remove" {
    value = step.pipeline.remove_user_from_group.output
  }
}



####    List     ####

pipeline "priv_command_list" {

  param "response_url" {}
  param username {}
  param channel_id {}

  step "query" "list_groups" {
    connection_string = "postgres://steampipe@127.0.0.1:9193/steampipe"
    sql = <<-EOQ
      select 
        -- name, 
        g ->> 'GroupName' as label
      from
        aws_iam_user,
        jsonb_array_elements(groups) as g
      where
        lower(name) = lower('${param.username}')
    EOQ
  }

  step "http" "respond" {
    description = "Respond to requestor with the results"
    url         = param.response_url
    method      = "post"

    request_body = jsonencode({
      text = <<-EOT
        %{ if length(step.query.list_groups.rows) == 0 }
        You are not a member of any groups.
        %{ else }
        You are a member of the following groups: 
        ${join(", ", step.query.list_groups.rows[*].label)}"
        %{ endif }
      EOT
    })
  }

  output "groups" {
      value = step.query.list_groups.rows
  }

}
