pipeline "nested_simple_function" {
  param "event" {
    default = {
        "message": "Hello, world!"
    }
  }

  step "function" "sleep" {
      source    = "./functions/nested_sleep"
      event     = param.event
      runtime = "nodejs:20"
  }

  output "val" {
    value = step.function.sleep
  }
}
