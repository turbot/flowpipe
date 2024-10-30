pipeline "for_loop" {

  param "users" {
    type    = list(string)
    default = ["jerry", "Janis", "Jimi"]
  }

  step "transform" "text_1" {
    for_each = param.users
    value    = "user if ${each.value}"
  }

  step "transform" "no_for_each" {
    value = "baz"
  }
}

pipeline "for_depend_object" {

  param "users" {
    type    = list(string)
    default = ["brian", "freddie", "john", "roger"]
  }

  step "transform" "text_1" {
    for_each = param.users
    value    = "user if ${each.value}"
  }

  step "transform" "text_3" {
    for_each = step.transform.text_1
    value    = "output one value is ${each.value.value}"
  }
}


pipeline "for_loop_depend" {

  param "users" {
    type    = list(string)
    default = ["jerry", "Janis", "Jimi"]
  }

  step "transform" "text_1" {
    for_each = param.users
    value    = "user is ${each.value}"
  }

  step "transform" "text_2" {
    value = "output is ${step.transform.text_1[0].value}"
  }

  step "transform" "text_3" {
    for_each = step.transform.text_1
    value    = "output one value is ${each.value.value}"
  }
}


pipeline "for_map" {
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
    value    = "${each.value.key} ${each.value.last_name} was ${each.value.age}"
  }
}
