pipeline "simple_loop" {
      
    step "transform" "repeat" {
        value  = "iteration"

        loop {
            value = loop.index + 1
        }
    }
}
