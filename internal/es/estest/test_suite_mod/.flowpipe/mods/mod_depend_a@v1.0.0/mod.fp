mod "mod_depend_a" {
    title = "Child mod A"

    require {
        mod "mod_depend_b" {
            version = "1.0.0"
        }
    }
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

pipeline "call_child_b" {
  step "pipeline" "child_mod" {
    pipeline = mod_depend_b.pipeline.echo_from_depend_b
  }

  output "out_from_b" {
    value = step.pipeline.child_mod.output.result
  }
}


pipeline "echo_b" {
  description = "this is from the var ${var.var_a_number}"
}