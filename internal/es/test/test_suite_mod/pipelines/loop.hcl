pipeline "simple_loop" {

    step "echo" "repeat" {
        text  = "iteration"
        numeric = 1

        loop {
            if = result.numeric > 5
            numeric = result.numeric + 1
        }
    }
}
