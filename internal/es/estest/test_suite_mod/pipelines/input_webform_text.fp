pipeline "input_webform_text" {

  step "input" "my_step" {
    type     = "text"
    prompt   = "Enter your name"
  }

  output "val" {
    value = step.input.my_step.value
  }

}