pipeline "function_step_test" {
  step "function" "my_func" {
    # Now
    source  = "./my-function"
    runtime = "nodejs"
    handler = "my_file.my_handler"
    timeout = 10

    env = {
      foo = "bar"
      up  = "down"
    }
  }
}

pipeline "function_step_test_with_param" {

  param "src" {
    type    = string
    default = "./my-function"
  }

  param "runtime" {
    type    = string
    default = "nodejs"
  }

  param "handler" {
    type    = string
    default = "my_file.my_handler"
  }

  param "timeout" {
    type    = number
    default = 10
  }

  step "function" "my_func" {
    # Now
    source  = param.src
    runtime = param.runtime
    handler = param.handler
    timeout = param.timeout

    env = {
      foo = "bar"
      up  = "down"
    }
  }
}

pipeline "function_step_test_string_timeout" {
  step "function" "my_func" {
    # Now
    source  = "./my-function"
    runtime = "nodejs"
    handler = "my_file.my_handler"
    timeout = "10s"

    env = {
      foo = "bar"
      up  = "down"
    }
  }
}

pipeline "function_step_test_string_timeout_with_param" {
  param "src" {
    type    = string
    default = "./my-function"
  }

  param "runtime" {
    type    = string
    default = "nodejs"
  }

  param "handler" {
    type    = string
    default = "my_file.my_handler"
  }

  param "timeout" {
    type    = string
    default = "10s"
  }

  step "function" "my_func" {
    # Now
    source  = param.src
    runtime = param.runtime
    handler = param.handler
    timeout = param.timeout

    env = {
      foo = "bar"
      up  = "down"
    }
  }
}
