pipeline "input_webform_select_with_labels_and_default_selection" {

  step "input" "select_region" {
    type     = "select"
    prompt   = "Select a region:"

    option "us-east-1" {
      label     = "N. Virginia"
      selected  = true
    }
    option "us-east-2" {
      label     = "Ohio"
    }
    option "us-west-2" {
      label     = "N. California"
    }
    option "us-west-2" {
      label     = "Oregon"
    }
  }

  output "val" {
    value = step.input.select_region.value
  }

}