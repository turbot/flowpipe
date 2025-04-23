mod "duplicate" {

}

pipeline "two_message_steps_with_same_name" {

  param "key" {
    type = number
    default = 1
  }

  step "message" "test" {
    if = param.key % 2 == 0
    notifier = notifier.default
    text = "Even: ${param.key}"
  }

  step "message" "test" {
    if = param.key % 2 == 1
    notifier = notifier.default
    text = "Odd: ${param.key}"
  }

}