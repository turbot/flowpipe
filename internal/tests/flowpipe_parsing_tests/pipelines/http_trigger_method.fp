pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "http" "trigger_without_method_block" {
  enabled  = true
  pipeline = pipeline.simple_with_trigger

  args = {
    param_one     = "one"
    param_two_int = 2
  }

  execution_mode = "synchronous"
}

trigger "http" "trigger_with_get_method" {
  enabled = true

  method "get" {
    pipeline = pipeline.simple_with_trigger

    args = {
      param_one     = "one"
      param_two_int = 2
    }

    execution_mode = "synchronous"
  }
}

trigger "http" "trigger_with_multiple_method" {
  enabled = true

  method "post" {
    pipeline = pipeline.simple_with_trigger

    args = {
      param_one     = "one"
      param_two_int = 2
    }

    execution_mode = "synchronous"
  }

  method "get" {
    pipeline = pipeline.simple_with_trigger

    args = {
      param_one     = "one"
      param_two_int = 3
    }

    execution_mode = "synchronous"
  }
}

trigger "http" "test_method_precedence" {
  enabled = true

  pipeline = pipeline.simple_with_trigger

  args = {
    param_one     = "one"
    param_two_int = 2
  }

  method "get" {
    pipeline = pipeline.simple_with_trigger

    args = {
      param_one     = "one"
      param_two_int = 3
    }

    execution_mode = "synchronous"
  }

}
