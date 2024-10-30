pipeline "validate_my_param" {

  param "my_token" {
    type    = string
    default = "token123"
  }

  param "my_number" {
    type    = number
    default = 234
  }

  param "my_number_two" {
    type    = number
    default = 123
  }

  param "my_bool" {
    type    = bool
    default = false
  }

  param "set_string" {
    type    = set(string)
    default = ["a", "b", "c"]
  }

  param "set_number" {
    type    = set(number)
    default = [1, 2, 3]
  }

  param "set_bool" {
    type    = set(bool)
    default = [true, false]
  }

  param "set_any" {
    type    = set
    default = ["a", "b", "c"]
  }

  param "list_string" {
    type    = list(string)
    default = ["a", "b", "c"]
  }

  param "list_bool" {
    type    = list(bool)
    default = [true, false]
  }

  param "list_number" {
    type    = list(number)
    default = [1, 2, 3]
  }

  param "list_number_two" {
    type    = list(number)
    default = [1, 2]
  }

  param "list_number_three" {
    type    = list(number)
    default = [1, 2, 3]
  }

  param "list_any" {
    type    = list
    default = ["a", "b", "c"]
  }

  param "list_any_two" {
    type    = list(any)
    default = ["a", "b", "c"]
  }

  param "list_any_three" {
    type    = list
    default = ["a", "b", "c"]
  }

  param "map_of_string" {
    type = map(string)
    default = {
      "a" = "b"
      "c" = "d"
    }
  }

  param "map_of_number" {
    type = map(number)
    default = {
      "a" = 1
      "b" = 2
    }
  }

  param "map_of_bool" {
    type = map(bool)
    default = {
      "a" = true
      "b" = false
    }
  }

  param "map_of_number_two" {
    type = map(number)
    default = {
      "a" = 1
      "b" = 2
    }
  }

  param "map_of_any" {
    type = map
    default = {
      "a" = "b"
      "c" = "d"
    }
  }

  param "map_of_any_two" {
    type = map
    default = {
      "a" = "b"
      "c" = "d"
    }
  }

  param "map_of_any_three" {
    type = map
    default = {
      "a" = "b"
      "c" = 3
    }
  }

  param "param_any" {
    type    = any
    default = "anything goes"
  }

  step "transform" "echo" {
    value = param.my_token
  }
}

