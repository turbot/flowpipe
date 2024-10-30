pipeline "test" {

    param "user_email" {
    }

    param "tokens" {
    }

    step "http" "http_test" {
        url = "https://localhost/index2.html"

        basic_auth {
            username = param.user_email
            password = param.tokens
        }
    }
}