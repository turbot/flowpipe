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

pipeline "json_decode_encode" {
    step "echo" "json_str" {
        text = jsonencode({
            Version = "2012-10-17"
            Users = ["jeff", "jerry", "jim"]
        })
    }

    step "echo" "json_obj" {
        json = jsondecode(step.echo.json_str.text)
    }
}


