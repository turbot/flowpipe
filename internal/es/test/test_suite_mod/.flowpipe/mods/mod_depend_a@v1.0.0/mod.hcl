mod "mod_depend_a" {
  title = "Child mod A"
}

variable "var_depend_a_one" {
  type = string
  default = "default value for var_depend_a_one variable"
}

variable "var_depend_a_two" {
  type = string

  # this is a bug, if we remove this default value in the child mod, Flowpipe doesn't start
  # workaround is to set a default in the child mod
  default = "no default"
}

pipeline "echo_one_depend_a" {
    step "echo" "echo_one" {
        text = "Hello World from Depend A"
    }

    step "echo" "var_one" {
        text = "Hello World from Depend A: ${var.var_depend_a_one}"
    }

    step "echo" "var_two" {
        text = "Hello World Two from Depend A: ${var.var_depend_a_two}"
    }

    step "echo" "echo_of_var_one" {
        text = "${step.echo.var_one.text} + ${var.var_depend_a_one}"
    }

    output "val" {
      value = step.echo.echo_one.text
    }

    output "val_var_one" {
      value = step.echo.echo_of_var_one.text
    }

    output "val_var_two" {
      value = step.echo.var_two.text
    }
}
