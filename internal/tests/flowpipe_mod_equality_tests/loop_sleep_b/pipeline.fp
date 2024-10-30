pipeline "sleep" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 3

           
        }
    }
}
