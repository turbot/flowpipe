pipeline "test" {

    param "foo" {
        type = string
    }
    
    step "http" "http_test" {
        url = "https://localhost/index.html"
    }

}