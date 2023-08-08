pipeline "bad_email_with_invalid_recipients" {

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

pipeline "bad_email_with_param" {

    param "from" {
      type    = string
      default = "sender@example.com"
    }

    param "sender_credential" {
      type    = string
      default = "sendercredential"
    }
    
    param "host" {
      type    = string
      default = "smtp.example.com"
    }

    param "port" {
      type    = string
      default = "587"
    }

    param "sender_name" {
      type    = string
      default = "Flowpipe"
    }

    step "email" "test_email" {
      to                = ["recipient@example.com"]
      from              = param.from
      sender_credential = param.sender_credential
      host              = param.host
      port              = param.port
      subject           = "Test email"
      body              = "This is a test email"
      sender_name       = param.sender_name
    }
}

pipeline "bad_email_with_expr" {

    param "to" {
      type    = list(string)
      default = ["recipient@example.com"]
    }
    
    param "sender_credential" {
      type    = string
      default = "sendercredential"
    }
    
    param "host" {
      type    = string
      default = "smtp.example.com"
    }

    param "port" {
      type    = string
      default = "587"
    }

    param "sender_name" {
      type    = string
      default = "flowpipe"
    }

    param "cc" {
      type    = list(string)
      default = ["ccrecipient@example.com"]
    }

    param "bcc" {
      type    = list(string)
      default = ["bccrecipient@example.com"]
    }

    step "echo" "sender_address" {
      text = "${param.sender_name}@example.com"
    }

    step "echo" "email_body" {
      text = "This is an email body"
    }

    step "email" "test_email" {
      to                = param.to
      from              = step.echo.sender_address.text
      sender_credential = param.sender_credential
      host              = param.host
      port              = param.port
      subject           = "Test email"
      body              = step.echo.email_body.text
      sender_name       = param.sender_name
      cc                = param.cc
      bcc               = param.bcc
    }
}
