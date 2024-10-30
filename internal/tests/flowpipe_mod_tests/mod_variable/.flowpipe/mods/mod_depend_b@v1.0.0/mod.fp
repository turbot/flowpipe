mod "mod_depend_b" {
  title = "Child mod B"
}

pipeline "echo_b" {
  description = "description from variable ${var.var_b_number}"
}