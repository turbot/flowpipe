pipeline "param_enum_test" {

  param "city" {
    type    = string
    default = "London"
    enum = ["London", "Paris", "New York"]
  }

  param "number" {
    type    = number
    default = 234
    enum = [234, 345, 456]
  }
}
