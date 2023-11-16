pipeline "jsonplaceholder_expr" {
    description = "Simple pipeline to demonstrate HTTP post operation."

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
            nested = {
                brian = {
                    may = "guitar"
                }
                freddie = {
                    mercury = "vocals"
                }
                roger = {
                    taylor = "drums"
                }
                john = {
                    deacon = "bass"
                }
            }
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
        json = step.http.http_1.response_body
    }

    step "echo" "body_json_nested" {
        text = step.http.http_1.response_body["nested"]["brian"]["may"]
    }

    step "echo" "body_json_loop" {
        for_each = step.http.http_1.response_body["title"]
        text = each.value
    }

    output "foo" {
        value = step.http.http_1.response_body
    }

    output "nested" {
        value = step.echo.body_json_nested.text
    }
}

pipeline "json_array" {

    param "request_body" {
        type = any
    }

    step "http" "json_http" {
        url = "https://jsonplaceholder.typicode.com/posts"

        method = "post"

        request_body = jsonencode(param.request_body)

        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
        }
    }


    step "echo" "json" {
        text = "[\"foo\", \"bar\", \"baz\"]"
    }

    output "val" {
        value = jsondecode(step.echo.json.text)
    }

    output "val_two" {
        value = step.http.json_http.response_body
    }
}