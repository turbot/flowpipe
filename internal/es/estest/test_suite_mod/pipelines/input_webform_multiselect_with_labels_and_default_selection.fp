pipeline "input_webform_multiselect_with_labels_and_default_selection" {

  step "input" "select_regions" {
    type     = "multiselect"
    prompt   = "Select regions:"

    option "us-east-1" {
      label     = "N. Virginia"
      selected  = true
    }
    option "us-east-2" {
      label     = "Ohio"
      selected  = true
    }

    option "us-west-2" {
      label     = "N. California"
    }
    option "us-west-2" {
      label     = "Oregon"
    }
  }

  output "val" {
    value = step.input.select_regions.value
  }

}