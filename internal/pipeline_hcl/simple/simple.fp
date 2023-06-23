pipeline "simple_http" {
    step "http" "my_step_1" {
        url = "http://localhost:8080"
    }
}