mod "test_mod" {
  title = "my_mod"
}

variable "var_one" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_one"
}

# var_two will be overriden in the test
variable "var_two" {
  type        = string
  description = "test variable"
  default = "default of var_two"
}


# var_three has no default
variable "var_three" {
  type        = string
  description = "test variable"
}


pipeline "one" {
    step "echo" "one" {
        text = "prefix text here and ${var.var_one} and suffix"
    }

    step "echo" "two" {
        text = "prefix text here and ${var.var_two} and suffix"
    }

    step "echo" "three" {
        text = "prefix text here and ${var.var_three} and suffix"
    }

    step "echo" "one_echo" {
        text =  "got prefix? ${step.echo.one.text} and again ${step.echo.one.text} and var ${var.var_one}"
    }
}
