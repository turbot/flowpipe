pipeline "pipeline_with_duplicate_param" {
  
  param "my_param" {
    type    = string
    default = "default"
  }

  param "my_param" {
    type    = number
    default = 10
  }

  step "transform" "print_param" {
    value = param.my_param
  }

  output "test" {
    value = step.transform.print_param.value
  }
}