pipeline "input_webform_multiselect_basic" {

  step "input" "select_regions" {
    notifier = notifier.default

    type     = "multiselect"
    prompt   = "Select regions:"

    option "us-east-1" {}
    option "us-east-2" {}
    option "us-west-1" {}
    option "us-west-2" {}

  }

  output "val" {
    value = step.input.select_regions.value
  }

}