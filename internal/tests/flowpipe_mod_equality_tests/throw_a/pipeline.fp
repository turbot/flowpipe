pipeline "test" {

    step "http" "http_test" {
        url = "https://localhost/index.html"

        throw {
            if = result.status_code == 400
            message = "HTTP status code is 400"
        }
    }
}