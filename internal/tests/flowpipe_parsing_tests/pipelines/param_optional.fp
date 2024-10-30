pipeline "test_param_optional" {
  param "test_optional" {
    type     = string
    optional = true
  }

  step "transform" "echo_optional" {
    if    = param.test_optional != null
    value = "optional but passed: ${param.test_optional}"
  }
  step "transform" "echo_optional_1" {
    if    = param.test_optional == null
    value = "optional and null"
  }

  output "test_output_1" {
    # value =  param.test_optional != null ? param.test_optional : "value was null"
    value = coalesce(param.test_optional, "value was null")
  }

}
