pipeline "onboard_employee" {
  title       = "Onboard Employee"
  description = "Onboard an employee"

  param "tools_needed" {
    type    = string
    default = "GITHUB"
  }

  step "transform" "check_github" {
    if    = contains(lower(param.tools_needed), "github")
    value = "contains github"
  }

  output "echo_check_gh" {
    value = step.transform.check_github
  }
}