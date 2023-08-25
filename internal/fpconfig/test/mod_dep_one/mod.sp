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
