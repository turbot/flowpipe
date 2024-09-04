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
  }

  step "transform" "creds" {
    value = credential.github[param.creds].token
  }

  step "transform" "merge_creds" {
    value = merge(credential.github[param.creds], {cred_name = param.creds})
  }

  output "val" {
    value = step.transform.creds.value
  }

  output "val_merge" {
    value = step.transform.merge_creds.value
  }
}

pipeline "with_github_conns" {
  param "conn" {
    type = string
  }

  step "transform" "conn" {
    value = connection.github[param.conn].token
  }

  step "transform" "merge_conns" {
    value = merge(connection.github[param.conn], {conn_name = param.conn})
  }

  output "val" {
    value = step.transform.conn.value
  }

  output "val_merge" {
    value = step.transform.merge_conns.value
  }
}
