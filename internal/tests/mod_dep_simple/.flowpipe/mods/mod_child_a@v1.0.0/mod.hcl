mod "mod_child_a" {
  title = "Child Mod A"
}

variable "var_one" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_one"
}

variable "var_two" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_two"
}

pipeline "this_pipeline_is_in_the_child" {
    step "echo" "foo" {
        text = "foo"
    }

    output "foo_a" {
        value = step.echo.foo.text
    }
}

pipeline "this_pipeline_is_in_the_child_using_variable" {
    step "echo" "foo" {
        text = "foo: ${var.var_one}"
    }
}

pipeline "this_pipeline_is_in_the_child_using_variable_passed_from_parent" {
    step "echo" "foo" {
        text = "foo: ${var.var_two}"
    }
}
