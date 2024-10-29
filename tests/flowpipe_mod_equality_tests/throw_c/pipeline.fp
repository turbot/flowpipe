pipeline "test" {

    step "http" "http_test" {
        url = "https://localhost/index.html"

        throw {
            if = result.status_code == 500
            message = "change here"
        }
    }
}