mod "test_mod" {
  title = "my_mod"
}


pipeline "echo" {
    step "transform" "echo" {
        value = var.slack_token
    }

    output "val" {
        value = step.transform.echo.value
    }
}
