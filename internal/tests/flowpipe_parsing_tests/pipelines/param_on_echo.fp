pipeline "param_on_echo" {
  param "test_optional" {
    type = string
    # optional = true
    # default = null
  }

  step "transform" "echo_one" {
    value = "Hello World"
  }

  step "transform" "test" {
    value = param.test_optional != null ? "${param.test_optional}" : "default"
  }

  output "echo_one_output" {
    value = step.transform.test.value
  }
}
