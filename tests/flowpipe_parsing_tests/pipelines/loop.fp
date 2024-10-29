pipeline "simple_loop" {

  step "transform" "repeat" {
    value = "iteration"

    loop {
      until = loop.index > 5
      value = loop.index + 1
    }
  }
}

pipeline "simple_http_loop" {

  step "http" "list_workspaces" {
    url    = "https://latestpipe.turbot.io/api/v1/org/latesttank/workspace/?limit=3"
    method = "get"

    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.pipes_token}"
    }

    loop {
      until = result.response_body.next_token != null
      url   = "https://latestpipe.turbot.io/api/v1/org/latesttank/workspace/?limit=3&next_token=${result.response_body.next_token}"
    }
  }
}

pipeline "pipeline_loop_test" {

  param "message" {
    type    = string
    default = "welcome"
  }

  param "index" {
    type = number
  }

  output "greet_world" {
    value = "Hello world! ${param.message} ${param.index}"
  }
}

pipeline "simple_pipeline_loop_unresolved" {

  param "test_message" {
    type    = string
    default = "hello"
  }

  step "pipeline" "repeat_pipeline_loop_test" {
    pipeline = pipeline.pipeline_loop_test
    args = {
      message = "from parent"
      index   = 0
    }

    loop {
      until = loop.index > 2
      args = {
        message = "${param.test_message}_${loop.index}"
        index   = loop.index
      }
    }
  }
}


pipeline "loop_depends_on_another_step" {

  step "sleep" "base" {
    duration = "5s"
  }

  step "sleep" "repeat" {
    duration = "iteration"

    loop {
      until    = loop.index > 5
      duration = step.sleep.base.duration + 1 + loop.index
    }
  }
}

pipeline "loop_resolved" {

  step "http" "repeat" {
    url = "https://does.not.matter"
    loop {
      until = try(result.response_body.next, null) == null
      url   = try(result.response_body.next, "")
    }
  }

}
