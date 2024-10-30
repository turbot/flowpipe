pipeline "if" {
  param "condition" {
    type    = bool
    default = true
  }

  step "transform" "text_1" {
    value = "foo"
    if    = param.condition
  }
}


pipeline "if_negative" {
  param "condition" {
    type    = bool
    default = true
  }

  step "transform" "text_1" {
    value = "foo"
    if    = param.condition
  }
}
