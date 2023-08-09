pipeline "email" {

    step "email" "test_email" {
      to                = ["recipient@example.com"]
      from              = "sender@example.com"
      sender_credential = "sendercredential"
      host              = "smtp.example.com"
      port              = "587"
      subject           = "Test email"
      body              = "This is a test email"
      sender_name       = "Flowpipe"
    }
}

pipeline "subscribe" {

  # param "subscriber" {
  #   default = ["recipient@example.com"]
  # }

  step "echo" "email_body" {
    text = "This is an email body"
  }

  step "email" "send_it" {
    to                = ["recipient@example.com"]
    from              = "sender@example.com"
    sender_credential = "sendercredential"
    host              = "smtp.example.com"
    port              = "587"
    subject           = "You have been subscribed"
    body              = step.echo.email_body.text
  }
}