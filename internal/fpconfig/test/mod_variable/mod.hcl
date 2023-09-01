mod "test_mod" {
  title = "my_mod"
}

variable "var_one" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_one"
}

pipeline "one" {
    step "echo" "one" {
        text = var.var_one
    }
}
