pipeline "test_basic_auth" {

    param "user_email" {
    }

    param "token" {
    }

    step "http" "http_test" {
        url = "http://localhost:7104/basic-auth-01"

        basic_auth {
            username = param.user_email
            password = param.token
        }
    }

    output "val" {
        value = step.http.http_test.response_body
    }
}