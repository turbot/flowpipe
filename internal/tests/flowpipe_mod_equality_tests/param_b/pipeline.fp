pipeline "test" {

    param "foo" {
        default = "bar"
    }
    
    step "http" "http_test" {
        url = "https://localhost/index.html"
    }

}