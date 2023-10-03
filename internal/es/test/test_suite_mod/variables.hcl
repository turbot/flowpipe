variable "var_one" {
  type        = string
  description = "test variable 2"
  default     = "this is the value of var_one"
}

# var_two has no default value
variable "var_two" {
  type        = string
  description = "test variable"
}

variable "var_three" {
  type        = string
  description = "test variable"
  default     = "if you see this then something is wrong"
}

variable "var_from_env" {
  type = string
  description = "will be set from env variable"
}

locals {
  locals_three_merge = merge(local.locals_three, {
    key_three = 33
  })
}

locals {
  locals_one = "value of locals_one"

  locals_two = 10

  locals_three = {
    key_one = "value of key_one"
    key_two = "value of key_two"
  }

  locals_four = ["foo", "bar", "baz"]
}
