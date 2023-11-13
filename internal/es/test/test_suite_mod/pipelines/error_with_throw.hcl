pipeline "error_with_throw_simple" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "bar"
        }

        retry {
            retries = 1
        }
    }
}