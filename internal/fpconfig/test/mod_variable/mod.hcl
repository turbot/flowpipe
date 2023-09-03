mod "test_mod" {
  title = "my_mod"
}

variable "var_one" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_one"
}

pipeline "one" {
    step "echo" "one" {
        text = "prefix text here and ${var.var_one} and suffix"
    }

    step "echo" "two" {
        text =  "got prefix? ${step.echo.one.text} and again ${step.echo.one.text} and var ${var.var_one}"
    }
}
