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

        insecure = param.insecure
    }

    step "transform" "output" {
        value = step.http.http_1.status_code
    }

    step "transform" "body_json" {
        value = step.http.http_1.response_body
    }

    step "transform" "body_json_nested" {
        value = step.http.http_1.response_body["nested"]["brian"]["may"]
    }

    step "transform" "body_json_loop" {
        for_each = step.http.http_1.response_body["title"]
        value = each.value
    }

    output "foo" {
        value = step.http.http_1.response_body
    }

    output "nested" {
        value = step.transform.body_json_nested.value
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


    step "transform" "json" {
        value = "[\"foo\", \"bar\", \"baz\"]"
    }

    output "val" {
        value = jsondecode(step.transform.json.value)
    }

    output "val_two" {
        value = step.http.json_http.response_body
    }

    output "val_request_body" {
        value = step.http.json_http.request_body
    }
}