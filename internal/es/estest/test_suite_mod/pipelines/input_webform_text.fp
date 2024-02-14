pipeline "input_webform_text" {

  step "input" "my_step" {
    type     = "text"
    prompt   = "Enter your name"
    option "A" {}
  }

}