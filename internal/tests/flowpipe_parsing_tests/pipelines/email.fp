pipeline "pipeline_with_email_step" {

  description = "Pipeline with a valid email step attributes"

  step "email" "valid_email" {

    // Require for the SMTP authentication
    smtp_username = "admiring.dijkstra@example.com"
    smtp_password = "abcdefghijklmnop"
    port          = 587
    host          = "smtp.gmail.com"

    // Sender and recipients
    from        = "sleepy.feynman@example.com"
    to          = ["friendly.curie@example.com", "angry.kepler@example.com"]
    sender_name = "Feynman" // optional

    // Email body
    subject      = "Flowpipe Test"                                                                 // optional
    body         = "This is a test plaintext email body to validate the email step functionality." // optional
    content_type = "text/plain"                                                                    // optional
    cc           = ["serene.turing@example.com"]                                                   // optional
    bcc          = ["elastic.bassi@example.com"]                                                   // optional
  }
}

pipeline "pipeline_with_unresolved_email_step_attributes" {

  description = "Pipeline with a valid email step attributes"

  param "smtp_username" {
    type    = string
    default = "admiring.dijkstra@example.com"
  }

  param "smtp_password" {
    type    = string
    default = "abcdefghijklmnop"
  }

  param "port" {
    type    = number
    default = 587
  }

  param "host" {
    type    = string
    default = "smtp.gmail.com"
  }

  param "from" {
    type    = string
    default = "sleepy.feynman@example.com"
  }

  param "sender_name" {
    type    = string
    default = "Feynman"
  }

  param "subject" {
    type    = string
    default = "Flowpipe Test"
  }

  param "content_type" {
    type    = string
    default = "text/plain"
  }

  param "body" {
    type    = string
    default = "This is a test plaintext email body to validate the email step functionality."
  }

  param "to" {
    type    = list(string)
    default = ["friendly.curie@example.com", "angry.kepler@example.com"]
  }

  param "cc" {
    type    = list(string)
    default = ["serene.turing@example.com"]
  }

  param "bcc" {
    type    = list(string)
    default = ["elastic.bassi@example.com"]
  }

  step "email" "valid_email" {

    // Require for the SMTP authentication
    smtp_username = param.smtp_username
    smtp_password = param.smtp_password
    port          = param.port
    host          = param.host

    // Sender and recipients
    from        = param.from
    to          = param.to
    sender_name = param.sender_name // optional

    // Email body
    subject      = param.subject      // optional
    body         = param.body         // optional
    content_type = param.content_type // optional
    cc           = param.cc           // optional
    bcc          = param.bcc          // optional
  }
}
