pipeline "sleep" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 3

            duration = "20s"           
        }
    }
}
