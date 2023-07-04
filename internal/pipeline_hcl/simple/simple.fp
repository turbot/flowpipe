pipeline "simple_http" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "http://localhost:8081"
    }

    step "sleep" "sleep_1" {
        duration = 20
    }

    step "email" "send_it" {
        to = "victor@turbot.com"
    }
}

pipeline "simple_http_2" {
    description = "my simple http pipeline 2"
    step "http" "my_step_1" {
        url = "http://localhost:8081"
    }
}