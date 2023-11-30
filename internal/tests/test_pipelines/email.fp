pipeline "email" {

    step "email" "test_email" {
      to                = ["recipient@example.com"]
      from              = "sender@example.com"
      smtp_password     = "sendercredential"
      smtp_username     = "sender@example.com"
      host              = "smtp.example.com"
      port              = 587
      subject           = "Test email"
      body              = "This is a test email"
      sender_name       = "Flowpipe"
    }
}

pipeline "subscribe" {

  # param "subscriber" {
  #   default = ["recipient@example.com"]
  # }

  step "transform" "email_body" {
    value = "This is an email body"
  }

  step "email" "send_it" {
    to                = ["recipient@example.com"]
    from              = "sender@example.com"
    smtp_password     = "sendercredential"
    smtp_username     = "sender@example.com"
    host              = "smtp.example.com"
    port              = 587
    subject           = "You have been subscribed"
    body              = step.transform.email_body.value
  }
}