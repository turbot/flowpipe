pipeline "simple_http" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "http://localhost:8081"
    }

    step "sleep" "sleep_1" {
        duration = "5s"
    }

    step "email" "send_it" {
        to                = ["victor@turbot.com"]
        from              = "sender@example.com"
        sender_credential = "sendercredential"
        host              = "smtp.example.com"
        port              = "587"
    }
}

pipeline "simple_http_2" {
    description = "my simple http pipeline 2"
    step "http" "my_step_1" {
        url = "http://localhost:8081"
    }
}


pipeline "sleep_with_output" {
    step "sleep" "sleep_1" {
        duration = "1s"
    }

    output "sleep_duration" {
      value = step.sleep.sleep_1.duration
    }
}
