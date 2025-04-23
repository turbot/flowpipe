pipeline "json" {
  step "transform" "json" {
    value = jsonencode({
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
  step "transform" "json" {
    value = jsonencode({
      Version = "2012-10-17"
      Users   = ["jeff", "jerry", "jim"]
    })
  }


  step "transform" "json_for" {
    for_each = step.transform.json.Users
    value    = "user: ${each.value}"
  }
}
