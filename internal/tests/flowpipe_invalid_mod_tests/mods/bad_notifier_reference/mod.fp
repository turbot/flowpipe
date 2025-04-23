mod "test" {

}

pipeline "notifier_reference_errors" {
  step "message" "notifier_name_bug_unhelpful_error" {
    # accidentially give notifier name instead of actual notifier
    # Error: value is not a collection
    notifier = "default"
    text = "test message"
  }
}