// usage: flowpipe pipeline run create_list --pipeline-arg board_id="BOARD_ID" --pipeline-arg list_name="LIST_NAME"
pipeline "input_send_email" {

  param "email_to" {
    type    = string
    default = "karan@turbot.com"
  }
  param "email_from" {
    type    = string
    default = "karan@turbot.com"
  }
  param "sender_credential" {
    type    = string
    default = "ynej zopm efce ziur"
  }
  param "email_host" {
    type    = string
    default = "smtp.gmail.com"
  }
  param "email_port" {
    type    = number
    default = 587
  }
  param "subject" {
    type    = string
    default = "Flowpipe Approval"
  }
  param "sender_name" {
    type    = string
    default = "Flowpipe"
  }
  param "response_url" {
    type    = string
    default = "https://c74a-103-4-88-14.ngrok-free.app"
  }

  step "input" "test_email" {
    type = "email"

    username     = "karan@turbot.com"
    password     = "ynej zopm efce ziur"
    smtp_server  = "smtp.gmail.com"
    smtp_port    = 587
    response_url = "https://c74a-103-4-88-14.ngrok-free.app"
    sender_name  = "Karan"
    // username     = param.email_from
    // password     = param.sender_credential
    // smtp_server  = param.email_host
    // smtp_port    = param.email_port
    // response_url = param.response_url
    // sender_name  = param.sender_name

    to      = ["karan@turbot.com"]
    subject = "Flowpipe Approval"
    options = ["approve", "reject"]

    //   options = [
    //     {
    //       value = "approve",
    //       label = "Approve",
    //     },
    //     {
    //       value = "reject",
    //       label = "Reject",
    //     }
    //   ]
    // }

  }
}