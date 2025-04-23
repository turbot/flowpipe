
variable "mandatory_tag_keys" {
  type        = list(string)
  description = "A list of mandatory tag keys to check for (case sensitive)."
  default     = ["Environment", "Owner"]
}


variable "var_number" {
  title = "variable with number default 42"
  type        = number
  default = 42
}

variable "var_map" {
    type = map(string)
    default = {
        key1 = "value1"
        key2 = "value2"
    }
}

variable "string_with_enum" {
    type = string
    enum = ["enum1", "enum2"]
    default = "enum1"
}

variable "number_with_enum" {
  type = number
  enum = [1, 2, 3]
  default = 1
}