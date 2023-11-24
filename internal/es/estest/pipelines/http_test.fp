pipeline "jsonplaceholder_expr" {
    description = "Simple pipeline to demonstrate HTTP post operation."

    step "transform" "method" {
        value = "post"
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

        method = step.transform.method.value

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

    step "transform" "output" {
        value = step.http.http_1.status_code
    }

    step "transform" "body_json" {
        value = step.http.http_1.response_body
    }

    step "transform" "body_json_loop" {
        for_each = step.http.http_1.response_body.title
        value    = each.value
    }

    output "foo" {
        value = step.http.http_1.response_body
    }
}
