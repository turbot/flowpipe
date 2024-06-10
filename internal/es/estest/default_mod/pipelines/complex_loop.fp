locals {
  use_message = "yes"
}

pipeline "caller" {
  step "pipeline" "step_name" {
    pipeline = pipeline.receiver
    args     = {
      a = [
        {
          name = "Bob"
          message = "some message here"
        },
        {
          name = "NotBob"
          message = "different message"
        }
      ]
    }
  }
}

pipeline "receiver" {
  param "a" {
    type = any
  }

  step "transform" "key_index" {
    value = {for item in param.a : item.name => item}
  }

    # this works
//   step "message" "example" {
//     if       = local.use_message == "yes"
//     for_each = step.transform.key_index.value
//     notifier = notifier["default"]
//     text     = "[Example] Key: ${each.key} -> ${each.value.message}"
//   }

    # this doesn't
  step "pipeline" "other" {
    if       = local.use_message != "yes"
    for_each = step.transform.key_index.value
    max_concurrency = 1
    pipeline  =  pipeline.receiver2
    args     = {
      item = each.value
    }
  }
}

pipeline "receiver2" {
  param "item" {
    type = any
  }

  step "message" "example" {
    notifier = notifier["default"]
    text     = "[Other] ${param.item.name}"
  }
}