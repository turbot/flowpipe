pipeline "all_param" {

    param "foo" {
        default = "bar"
    }

    step "echo" "echo" {
        text = param.foo
    }

   step "echo" "echo_foo" {
        text = "${param.foo}"
    }

    step "echo" "echo_three" {
        text = "${step.echo.echo.text} and ${param.foo}"
    }

    step "echo" "echo_baz" {
        text = "foo"
    }
}