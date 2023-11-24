mod "mod_depend_a" {
  title = "Child mod A"
}

pipeline "echo_one_depend_a" {
    step "transform" "echo_one" {
        value = "Hello World from Depend A"
    }

    output "val" {
      value = step.transform.echo_one.value
    }
}
