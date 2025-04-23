pipeline "test" {

    param "foo" {
        default = "foo"
    }
    
    step "http" "http_test" {
        url = "https://localhost/index.html"
    }

}