
mod "local" {
  title = "my_mod"
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
}

pipeline "json_for" {
    step "echo" "json" {
        json = jsonencode({
            Version = "2012-10-17"
            Users = ["jeff", "jerry", "jim"]
        })
    }


    step "echo" "json_for" {
        for_each = step.echo.json.Users
        text = "user: ${each.value}"
    }
}