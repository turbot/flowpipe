pipeline "retry_invalid_attribute_for_strategy" {

    step "transform" "one" {
        value = "foo"

        retry {
            strategy = "foo"            
        }
    }
}
