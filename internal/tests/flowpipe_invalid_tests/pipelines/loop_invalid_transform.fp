pipeline "simple_loop" {

  step "transform" "repeat" {
    value = "iteration"

    loop {
      until = loop.index > 5
      baz   = loop.index + 1
    }
  }
}
