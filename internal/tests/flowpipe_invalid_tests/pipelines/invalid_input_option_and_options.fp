pipeline "pipeline_with_option_and_options" {
  step "input" "multiple_option_and_options" {
    notifier = notifier.default
    
    type   = "button"
    prompt = "choose one:"

    option "yes" {}
    option "maybe" {}

    options = [
      {
        "value": "no"
      }
    ]
  }
}