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
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_a" {
        value = step.transform.foo.value
    }
}

pipeline "this_pipeline_is_in_the_child_using_variable" {
    step "transform" "foo" {
        value = "foo: ${var.var_one}"
    }
}

pipeline "this_pipeline_is_in_the_child_using_variable_passed_from_parent" {
    step "transform" "foo" {
        value = "foo: ${var.var_two}"
    }
}
