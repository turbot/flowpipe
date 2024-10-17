mod "mod_depend_b" {
  title = "Child Mod B"
}

pipeline "echo_from_depend_b" {
  step "transform" "echo" {
    value = "Hello World from Depend B"
  }

  output "result" {
    value = step.transform.echo.value
  }
}