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

// Important to leave this pipeline here. There was a bug that creds in the nested pipelines are not
// resolved correctly
pipeline "with_github_creds" {
  param "creds" {
    type = string
    default = "default"
  }

  step "transform" "creds" {
    value = credential.github[param.creds].token
  }
}
