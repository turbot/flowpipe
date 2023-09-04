mod "mod_depend_a" {
  title = "Child mod A"
}


pipeline "echo_one_depend_a" {
    step "echo" "echo_one" {
        text = "Hello World from Depend A"
    }
}
