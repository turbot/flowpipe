pipeline "simple_loop" {

    step "echo" "repeat" {
        text  = "iteration: ${loop.index}"
        numeric = 1

        loop {
            if = result.numeric < 3
            numeric = result.numeric + 1
        }
    }
}
