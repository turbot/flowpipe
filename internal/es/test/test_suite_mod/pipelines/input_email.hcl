// usage: flowpipe pipeline run input_send_email
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
    default = "xxx"
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
    default = "https://4d0c-103-4-88-14.ngrok-free.app"
  }

  step "input" "test_email" {
    type = "email"

    username     = param.email_from
    password     = param.sender_credential
    smtp_server  = param.email_host
    smtp_port    = param.email_port
    response_url = param.response_url
    sender_name  = param.sender_name

    to      = ["${param.email_to}"]
    subject = param.subject
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