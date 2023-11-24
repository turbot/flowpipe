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
    step "transform" "echo_one" {
        value = "Hello World from Depend A"
    }

    step "transform" "var_one" {
        value = "Hello World from Depend A: ${var.var_depend_a_one}"
    }

    step "transform" "var_two" {
        value = "Hello World Two from Depend A: ${var.var_depend_a_two}"
    }

    step "transform" "echo_of_var_one" {
        value = "${step.transform.var_one.value} + ${var.var_depend_a_one}"
    }

    output "val" {
      value = step.transform.echo_one.value
    }

    output "val_var_one" {
      value = step.transform.echo_of_var_one.value
    }

    output "val_var_two" {
      value = step.transform.var_two.value
    }
}
