pipeline "retry_multiple_retry_blocks" {

    step "transform" "one" {
        value = "foo"

        retry {
            max_attempts = "foo"
        }
    }
}
