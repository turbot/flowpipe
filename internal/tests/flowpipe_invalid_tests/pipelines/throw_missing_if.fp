pipeline "throw_invalid_attribute" {

    step "transform" "one" {
        value = "foo"

        throw {
            message = "bar"
        }
    }
}
