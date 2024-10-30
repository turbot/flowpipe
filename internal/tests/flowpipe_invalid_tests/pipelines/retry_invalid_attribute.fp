pipeline "retry_invalid_attribute" {

    step "transform" "one" {
        value = "foo"

        retry {
            max_attempts = 3
            except = "foo"
        }
    }
}
