mod "mod_depend_a" {
  title = "Child mod A"
}

pipeline "echo_one_depend_a" {
    step "transform" "echo_one" {
        value = "Hello World from Depend A"
    }

    output "val" {
      value = step.transform.echo_one.value
    }
}

pipeline "with_github_creds" {
  param "creds" {
    type = string
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


trigger "schedule" "http_step_trigger_in_a" {
  schedule = "* * * * *"

  enabled = false

  pipeline = pipeline.http
}