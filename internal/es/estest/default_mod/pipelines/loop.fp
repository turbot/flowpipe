pipeline "simple_pipeline_loop_with_args" {

  param "test_message" {
    type    = string
    default = "loop index"
  }

  step "pipeline" "repeat_pipeline_loop_test" {
    pipeline = pipeline.pipeline_loop_test
    args = {
      message = "iteration index"
      index   = 0
    }

    loop {
      // until = loop.index > 5
      until = try(result.output.greet_world, null) == null
      args = {
        message = "${param.test_message}_${loop.index}"
        index   = loop.index
      }
    }
  }

  output "value" {
    value = step.pipeline.repeat_pipeline_loop_test
  }
}

pipeline "pipeline_loop_test" {

  param "message" {
    type = string
  }

  param "index" {
    type = number
  }

  output "greet_world" {
    value =  param.index < 2 ?  "Hello world! ${param.message} ${param.index}" : null
  }
}
