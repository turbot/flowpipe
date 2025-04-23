mod "mod_depend_a_1" {
  title = "Child mod A.1"
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


trigger "schedule" "http_step_trigger_in_a1" {
  schedule = "* * * * *"

  enabled = false

  pipeline = pipeline.http
}