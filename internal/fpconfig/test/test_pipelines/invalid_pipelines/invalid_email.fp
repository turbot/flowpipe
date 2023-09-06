pipeline "invalid_port" {
  description = "invalid smtp port format"

  step "email" "invalid_smtp_port" {
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