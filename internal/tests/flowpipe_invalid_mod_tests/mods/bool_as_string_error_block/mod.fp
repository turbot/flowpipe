mod "test" {

}

pipeline "retry_with_linear_backoff" {

    step "transform" "one" {
        value = "foo"

        error {
            ignore = "true"
        }

    }
}
