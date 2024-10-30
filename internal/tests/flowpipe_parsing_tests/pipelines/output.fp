pipeline "with_output" {
  step "transform" "echo_1" {
    value = "foo bar"
  }

  output "one" {
    value = step.transform.echo_1.value
  }

  output "two" {
    value = title(step.transform.echo_1.value)
  }
}
