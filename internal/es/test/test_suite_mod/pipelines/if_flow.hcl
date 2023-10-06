pipeline "if_flow" {

    param "run_bad_step" {
        type = bool
        default = false
    }

    param "data" {
        type = any
        default = {
            "foo" = {
                "bar" = {
                    "baz" = "qux"
                }
            }
        }
    }

    step "echo" "echo" {
        text = param.data.foo.bar.baz
    }

    step "echo" "run_bad_step" {
        if = param.run_bad_step
        text = param.data.foo.bar.bad
    }

    output "echo"  {
        value = step.echo.echo.text
    }

    output "echo_bad" {
        value = step.echo.run_bad_step.text
    }
}