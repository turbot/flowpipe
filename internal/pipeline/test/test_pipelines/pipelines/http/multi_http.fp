pipeline "multi_http" {
    step "http" "my_step_1" {
        url = "https://example.com/my/webhook1"
    }
    step "http" "my_step_2" {
        url = "https://example.com/my/webhook1"
    }
}