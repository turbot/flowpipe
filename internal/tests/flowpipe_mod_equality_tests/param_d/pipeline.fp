pipeline "test" {

    param "foo" {
        type = number
    }
    
    step "http" "http_test" {
        url = "https://localhost/index.html"
    }

}