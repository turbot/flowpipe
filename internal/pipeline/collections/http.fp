pipeline "simple_http" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.json"
    }
}

pipeline "jsonplaceholder" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "https://jsonplaceholder.typicode.com/posts"
        method = "Post"
        request_body = jsonencode({
            userId = 12345
            title = "brian may"
        })
        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
            User-Agent = "flowpipe"
        }
        request_timeout_ms = 3000
    }

    step "echo" "output" {
        text = step.http.my_step_1.status_code
    }
}


pipeline "jsonplaceholder_expr" {
    description = "my simple http pipeline"

    step "echo" "method" {
        text = "post"
    }

    param "timeout" {
        type = number
        default = 1000
    }

    step "http" "http_1" {
        url = "https://jsonplaceholder.typicode.com/posts"

        method = step.echo.method.text

        request_body = jsonencode({
            userId = 12345
            title = "brian may"
        })
        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
            User-Agent = "flowpipe"
        }
        # this doesn't work yet
        # request_timeout_ms = param.timeout.value
    }

    step "echo" "output" {
        text = step.http.http_1.status_code
    }
}
