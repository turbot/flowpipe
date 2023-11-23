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

    step "transform" "echo" {
        value = param.data.foo.bar.baz
    }

    step "transform" "run_bad_step" {
        if    = param.run_bad_step
        value = param.data.foo.bar.bad
    }

    output "echo"  {
        value = step.transform.echo.value
    }

    output "echo_bad" {
        value = step.transform.run_bad_step.value
    }
}