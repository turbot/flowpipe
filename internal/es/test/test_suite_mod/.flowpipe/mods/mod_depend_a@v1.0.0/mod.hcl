mod "mod_depend_a" {
  title = "Child mod A"
}

variable "var_depend_a_one" {
  type = string
  default = "default value for var_depend_a_one variable"

}

pipeline "echo_one_depend_a" {
    step "echo" "echo_one" {
        text = "Hello World from Depend A"
    }

    step "echo" "var_one" {
        text = "Hello World from Depend A: ${var.var_depend_a_one}"
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
}
