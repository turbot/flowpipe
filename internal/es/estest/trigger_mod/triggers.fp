trigger "schedule" "report_trigger" {
  schedule = "* * * * *"

  enabled = true

  param "param_one" {
    type = string
    default = "value_one"
  }

  param "param_two" {
    type = string
    default = "value_two"
  }

  param "param_three" {
    type = number
    default = 42
  }

  param "param_four" {
    type = map(string)
    default = {
      "foo": "bar"
      "bar": "baz"
    }
  }

  param "param_five" {
    type = map(number)
    default = {
      "foo": 1
      "bar": 2
    }
  }

  pipeline = pipeline.report_pipeline

  args = {
    gh_repo = param.param_one
  }
}


pipeline "report_pipeline" {
  param "gh_repo" {
    type    = string
    default = "bar"
  }

  step "transform" "echo" {
    value = param.gh_repo
  }

  output "val" {
    value = step.transform.echo.value
  }
}

pipeline "http" {
    description = "Bad HTTP step, just one step in the pipeline."

    step "http" "my_step_1" {
        url = "https://www.google.coms"
    }

    output "http" {
      value = step.http.my_step_1.response_body
    }
}



trigger "schedule" "http_step_trigger" {
  schedule = "* * * * *"

  enabled = false

  pipeline = pipeline.http
}