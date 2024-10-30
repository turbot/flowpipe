pipeline "all_param" {

  param "foo" {
    default = "bar"
  }

  step "transform" "echo" {
    value = param.foo
  }

  step "transform" "echo_foo" {
    value = param.foo
  }

  step "transform" "echo_three" {
    value = "${step.transform.echo.value} and ${param.foo}"
  }

  step "transform" "echo_baz" {
    value = "foo"
  }
}
