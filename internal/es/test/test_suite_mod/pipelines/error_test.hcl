pipeline "validate_error" {

  step "pipeline" "pipeline_step" {
    pipeline = pipeline.execute_http
  }

  output "pipeline_step_output" {
    value = step.pipeline.pipeline_step.output.foo
  }
}

pipeline "execute_http" {

  step "http" "http_step" {
    url    = "https://google.com/foobar.json"
    method = "get"
  }

  output "foo" {
    value = step.http.http_step.status_code
  }
}
