pipeline "loop_sleep" {

    step "sleep" "sleep" {
        duration = "1s"

        loop {
            until = loop.index > 2
            duration = "${loop.index}s"
        }
    }
}