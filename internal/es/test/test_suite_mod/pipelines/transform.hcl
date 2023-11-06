pipeline "pipeline_with_transform_step" {

  description = "Pipeline with transform step"

  param  "transform_value_ref" {
    type = number
    default = 10
  }

  step "transform" "basic_transform" {
    value = "This is a simple transform step"
  }

  step "transform" "basic_transform_refers_param" {
    value = param.transform_value_ref
  }

  step "transform" "depends_on_transform_step" {
    value = "${step.transform.basic_transform.value} - test123"
  }
}

pipeline "pipeline_with_transform_step_string_list" {

  description = "Pipeline with transform step contains list of strings"

  param "users" {
    type    = list(string)
    default = ["brian", "freddie", "john", "roger"]
  }

   step "transform" "transform_test" {
    for_each = param.users
    value    = "user if ${each.value}"
  }
}