mod "test" {

}

pipeline "retry_with_linear_backoff" {

    step "transform" "one" {
        value = "foo"

        retry { 
            strategy = "linear"
            min_interval = "500"
            max_interval = 4000
        }

    }
}
