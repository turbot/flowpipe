pipeline "http_step" {
    step "http" "send_to_slack" {
    url                = "https://myapi.com/vi/api/do-something"
    method             = "post"
    insecure           = true
    // ca_cert_pem        = file("${path.module}/certs/CA_crt.pem")
    request_timeout_ms = 2000

    request_body = jsonencode({
        name =  "turbie"
        app  =  "flowpipe"
    })

    request_headers = {
      Accept      = "application/json"
      User-Agent  = "flowpipe"   // check - is this the syntax with dash in a key name???
    }

    }
}

pipeline "bad_http" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"

        error {
          ignore = true
        }
    }

    step "echo" "bad_http" {
        for_each = step.http.my_step_1.errors
        text = each.message
    }
}
