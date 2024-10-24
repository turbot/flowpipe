variable "var_number" {
  type        = number
  default = 42
  enum = [42, 43]
  tags = {
    "Environment" = "dev"
    "Owner" = "me"
  }
}


variable "var_number_list" {
  type        = list(number)
  default = [1, 2, 3]
  tags = {
    "Environment" = "dev"
    "Owner" = "me"
  }
}

variable "var_string" {
  type        = string
  default = "default"
  enum = ["default", "other"]
  tags = {
    "Environment" = "dev"
    "Owner" = "me"
  }
}

variable "complex_var" {
  type = map(object({
    name = string
    age = number
  }))
  default = {
    "first" = {
      name = "John"
      age = 42
    }
    "second" = {
      name = "Jane"
      age = 43
    }
  }
}