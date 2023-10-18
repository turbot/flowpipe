
trigger "http" "priv_cmd" {
    title    = "Webhook Trigger for Slack /priv command"
    pipeline = pipeline.priv_command_router
    args     = {
      response_url  = parse_query_string(self.request_body).response_url
      slack_text    = parse_query_string(self.request_body).text
      body          = self.request_body
    }
}



pipeline "priv_command_router" {

  param "response_url" {}
  param "slack_text" {}

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
  }

  step "pipeline" "add" {
    if       = length(step.echo.parse.output.command) > 0 && step.echo.parse.output.command[0] == "add"
    pipeline = pipeline.priv_command_add

    args = {
      username   = step.echo.parse.output.requestor
      channel_id = step.echo.parse.output.channel_id
    }
    
  }

}

pipeline "priv_command_add" {

  // param "response_url" {}
  // param "slack_text" {}

  param username {}
  param channel_id {}

  param "body" {}

  // step "echo" "parse" {
  //   text = param.body

  //   output "requestor" {
  //     value = parse_query_string(param.body).user_name
  //   }

  //   output "command" {
  //     value = split("+", parse_query_string(param.body).text)
  //   }

  //   output "channel_id" {
  //     value = parse_query_string(param.body).channel_id
  //   }
  // }


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
    token      = var.slack_token
    channel    = param.channel_id # "#mantix_test" #step.echo.parse.output.channel_id
    slack_type = "button"
    prompt     = "Select a group"
    options    = step.query.list_groups.rows[*].label
  }

  step "pipeline" "add_user_to_group" {

    pipeline = pipeline.user_group_add

    args = {
      username = param.username #step.echo.parse.output.requestor
      group    = step.input.select_group.value
    }

  }


  // step "http" "respond" {
  //   #description = "Respond to slack with the acknowledgement"
  //   url         = param.response_url
  //   method      = "post"

  //   // request_body = jsonencode({
  //   //   text = join(", ", step.query.list_groups.rows[*].label)
  //   // })

  //   request_body = jsonencode({
  //     text = param.body
  //   })

  //   // request_body = jsonencode({
  //   //   text = step.query.list_groups.rows[0].name
  //   // })
  //   //request_body = jsonencode({
  //   //  text = "ok"
  //   //})
  // }



  // step "input" "approval" {
  //   type       = "slack"
  //   token      = var.slack_token
  //   channel    = step.echo.parse.channel_id
  //   slack_type = "button"
  //   prompt     = "User ${step.echo.parse.output.requestor} has requested to be added to the ${step.input.select_group.value} group. Do you want to approve?"
  //   options    = ["Approve","Deny"]
  // }

  output "groups" {
      value = step.query.list_groups.rows
  }

  output "user_group_add" {
    value = step.pipeline.add_user_to_group.output
  }

}
