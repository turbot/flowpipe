variable "var_number" {
  type        = number
  default = 42
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
