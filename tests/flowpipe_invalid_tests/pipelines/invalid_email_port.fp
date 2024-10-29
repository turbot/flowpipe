pipeline "invalid_port" {
  description = "invalid smtp port format"

  step "email" "invalid_smtp_port" {
    to            = ["recipient@example.com"]
    from          = "sender@example.com"
    smtp_password = "sendercredential"
    smtp_username = "sender@example.com"
    host          = "smtp.example.com"
    port          = "587"
    subject       = "Test email"
    body          = "This is a test email"
    sender_name   = "Flowpipe"
  }
}
