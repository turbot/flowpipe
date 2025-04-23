pipeline "simple_loop" {

  step "sleep" "repeat" {
    duration = "5s"

    loop {
      until = loop.index > 5
      baz   = loop.index + 1
    }
  }
}
