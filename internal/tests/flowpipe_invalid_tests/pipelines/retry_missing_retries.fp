pipeline "retry_missing_retries" {

    step "transform" "one" {
        value = "foo"

        retry {
            if = result.value == "foo"
        }
    }
}
