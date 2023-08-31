mod "mod_parent" {
  title = "Parent Mod"
  require {
    mod "mod_child_a" {
        version = "1.0.0"
    }
  }
}

pipeline "json" {
    step "echo" "json" {
        json = jsonencode({
            Version = "2012-10-17"
            Statement = [
            {
                Action = [
                "ec2:Describe*",
                ]
                Effect   = "Allow"
                Resource = "*"
            },
            ]
        })
    }

    output "foo" {
        value = step.echo.json.json
    }
}

// pipeline "refer_to_child" {
//     step "pipeline" "child_output" {
//         pipeline = mod_child_a.pipeline.this_pipeline_is_in_the_child
//     }
// }
