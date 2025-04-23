
pipeline "if_loop" {

  param "messages" {
    type = list(string)
    default = ["a", "b", "c"]
  }

  # loop is evaluated last, since there's if = false, this step's loop will never be executed
  # it should only have 1 step execution which is the one to evaluate the if condition
  step "message" "test" {
    notifier = notifier["default"]
    if = false
    loop {
      until = loop.index == length(param.messages)-1
    }
    text = param.messages[loop.index]
  }
}
