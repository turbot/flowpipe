pipeline "simple" {
    step "echo" "echo" {
        text = "Hello World"

        output "echo_1" {
            value = "echo 1"
        }

        output "echo_2" {
            value = "echo 2"
        }
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

pipeline "simple_for_each" {

    step "transform" "echo" {
        for_each = ["bar", "baz", "qux"]

        value = "${each.key}: foo ${each.value}"

        output "val" {
            value = "val is: ${each.value}"
        }
    }

    output "val" {
        value = step.transform.echo
    }
}