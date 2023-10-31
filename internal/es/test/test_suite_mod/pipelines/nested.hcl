pipeline "top" {

    step "echo" "hello" {
        text = "hello world"
    }

    step "pipeline" "middle" {
        pipeline = pipeline.middle

        args = {
            issue_title = "hello world"
        }
    }


    step "echo" "combine" {
        text = step.pipeline.middle.output.val
    }

    output "val" {
        value = step.echo.combine.text
    }

    output "val_two" {
        value = step.pipeline.middle.output.val_two
    }
}

pipeline "middle" {

    param "issue_title" {
        type = string
    }

    step "echo" "echo" {
        text = "middle world"
    }

    step "pipeline" "call_bottom" {
        pipeline = pipeline.bottom
    }

    step "echo" "echo_two" {
        json = jsonencode({
          query = <<EOQ
            mutation {
                createIssue(input: {repositoryId: "${step.pipeline.call_bottom.output.repository_id}", title: "${param.issue_title}"}
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
        value = step.echo.echo.text
    }

    output "val_two" {
        value = step.echo.echo_two.json
    }
}


pipeline "bottom" {


    step "echo" "echo" {
        json = jsonencode({
            jerry = "garcia"
            jimmy = "hendrix"
        })
    }

    output "val" {
        value = step.echo.echo.json
    }

    output "repository_id" {
        value = jsondecode(step.echo.echo.json).jimmy
    }
}