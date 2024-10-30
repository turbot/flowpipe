
pipeline "retry_with_default_backoff" {

    step "transform" "one" {
        value = "foo"

        retry { }
    }
}


pipeline "retry_with_linear_backoff" {

    step "transform" "one" {
        value = "foo"

        retry { 
            strategy = "linear"
            min_interval = 500
            max_interval = 4000
        }
    }
}

pipeline "retry_with_exponential_backoff" {

    step "transform" "one" {
        value = "foo"

        retry { 
            strategy = "exponential"
            min_interval = 500
            max_interval = 50000
        }
    }
}