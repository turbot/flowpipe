pipeline "retry_default" {

    step "transform" "one" {
        value = "foo"

        retry { 

        }
    }
}

pipeline "retry_simple" {

    step "transform" "one" {
        value = "foo"

        retry {
            max_attempts = 2
            strategy = "exponential"
        }
    }
}

pipeline "retry_with_if" {

    step "transform" "one" {
        value = "foo"

        retry {
            if = result.value == "foo"
            max_attempts = 5
        }
    }
}