pipeline "test" {

    step "http" "http_test" {
        url = "https://localhost/index.html"
    }

}