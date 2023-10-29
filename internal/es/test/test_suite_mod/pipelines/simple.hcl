pipeline "simple" {
    step "echo" "echo" {
        text = "Hello World"
    }

    output "val" {
        value = step.echo.echo.text
    }
}

pipeline "simple_two_steps" {

    step "echo" "echo" {
        text = "Hello World"
    }

    step "echo" "echo_two" {
        text = "${step.echo.echo.text}: Hello World"
    }

    output "val" {
        value = step.echo.echo.text
    }

    output "val_two" {
        value = step.echo.echo_two.text
    }
}