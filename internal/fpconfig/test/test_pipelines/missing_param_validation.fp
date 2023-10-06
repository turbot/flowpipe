pipeline "missing_param_validation_test" {
  
  param "address_line_1" {
    type = string
    default = "10 Downing Street"
  }

  param "address_line_2" {
    type = string
  }

  param "city" {
    type = string
    default = "London"
  }

  step "echo" "greetings" {
    text = "Hello, welcome to ${param.address_line_1}, ${param.address_line_2}, ${param.city}"
  }

  output "greetings_text" {
    value = step.echo.greetings.text
  }
}
