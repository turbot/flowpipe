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

  step "transform" "number" {
    value = 23
  }

  output "basic_transform" {
    value = step.transform.basic_transform.value
  }

  output "depends_on_transform_step" {
    value = step.transform.depends_on_transform_step.value
  }

  output "number" {
    value = step.transform.number.value

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
    value    = "user is ${each.value}"
  }

  output "transform_test" {
    value = step.transform.transform_test
  }
}

pipeline "transform_step_for_map" {
  param "legends" {
    type = map

    default = {
      "janis" = {
        last_name = "joplin"
        age       = 27
      }
      "jimi" = {
        last_name = "hendrix"
        age       = 27
      }
      "jerry" = {
        last_name = "garcia"
        age       = 53
      }
    }
  }

  step "transform" "text_1" {
    for_each = param.legends
    value    = "${each.key} ${each.value.last_name} was ${each.value.age}"
  }

  output "text_1" {
    value = step.transform.text_1
  }
}

pipeline "transform_loop" {
    step "transform" "foo" {
        value = "loop: ${loop.index}"

        loop {
            until = loop.index >= 2
            value = "loop: ${loop.index}"
        }
    }

    output "val" {
        value = step.transform.foo
    }

    output "val_1" {
        value = step.transform.foo[0].value
    }
    output "val_2" {
        value = step.transform.foo[1].value
    }
    output "val_3" {
        value = step.transform.foo[2].value
    }
}
