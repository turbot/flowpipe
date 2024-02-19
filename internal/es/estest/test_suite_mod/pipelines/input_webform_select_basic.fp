pipeline "input_webform_select_basic" {

  step "input" "select_region" {
    type     = "select"
    prompt   = "Select a region:"

    option "us-east-1" {}
    option "us-east-2" {}
    option "us-west-1" {}
    option "us-west-2" {}

  }

  output "val" {
    value = step.input.select_region.value
  }

}