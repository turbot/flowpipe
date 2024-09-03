variable "var_number" {
  type        = number
  default = 42
  tags = {
    "Environment" = "dev"
    "Owner" = "me"
  }
}
