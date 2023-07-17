pipeline "simple_sleep" {
    description = "my simple sleep pipeline"
    step "sleep" "sleep_1" {
        duration = "1s"
    }
}