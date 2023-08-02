pipeline "jsonplaceholder_expr" {
    description = "my simple http pipeline"

    step "echo" "method" {
        text = "post"
    }

    param "timeout" {
        type = number
        default = 1000
    }

    param "user_agent" {
        type = string
        default = "flowpipe"
    }

    param "insecure" {
        type = bool
        default = true
    }

    step "http" "http_1" {
        url = "https://jsonplaceholder.typicode.com/posts"

        method = step.echo.method.text

        request_body = jsonencode({
            userId = 12345
            title = ["brian may", "freddie mercury", "roger taylor", "john deacon"]
        })

        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
            User-Agent = param.user_agent
        }

        request_timeout_ms = param.timeout

        insecure = param.insecure
    }

    step "echo" "output" {
        text = step.http.http_1.status_code
    }

    step "echo" "body_json" {
        json = jsondecode(step.http.http_1.response_body)
    }

    step "echo" "body_json_loop" {
        for_each = jsondecode(step.http.http_1.response_body).title
        text = each.value
    }

    output "foo" {
        value = step.http.http_1.response_body
    }
}
