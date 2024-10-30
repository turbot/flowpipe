pipeline "top" {

  step "transform" "hello" {
    value = "hello world"
  }

  step "pipeline" "middle" {
    pipeline = pipeline.middle

    args = {
      issue_title = "hello world"
    }
  }


  step "transform" "combine" {
    value = step.pipeline.middle.val
  }

  output "val" {
    value = step.transform.combine.value
  }
}

pipeline "middle" {

  param "issue_title" {
    type = string
  }

  step "transform" "echo" {
    value = "middle world"
  }

  step "pipeline" "call_bottom" {
    pipeline = pipeline.bottom
  }

  step "transform" "echo_two" {
    value = jsonencode({
      query = <<EOQ
            mutation {
                createIssue(input: {repositoryId: "${step.pipeline.call_bottom.repository_id}", title: "${param.issue_title}"}
                ) {
                    clientMutationId
                    issue {
                      id
                      url
                    }
                }
            }
            EOQ
    })
  }


  output "val" {
    value = step.transform.echo.value
  }

  output "val_two" {
    value = step.transform.echo_two.value
  }
}


pipeline "bottom" {


  step "transform" "echo" {
    value = jsonencode({
      jerry = "garcia"
      jimmy = "hendrix"
    })
  }

  output "val" {
    value = step.transform.echo.value
  }

  output "repository_id" {
    value = jsondecode(step.transform.echo.value).jimmy
  }
}
