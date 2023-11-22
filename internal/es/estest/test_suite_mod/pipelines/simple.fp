pipeline "simple" {
    step "transform" "echo" {
        value = "Hello World"

        output "echo_1" {
            value = "echo 1"
        }

        output "echo_2" {
            value = "echo 2"
        }
    }

    output "val" {
        value = step.transform.echo.value
    }
}

pipeline "simple_two_steps" {

    step "transform" "echo" {
        value = "Hello World"
    }

    step "transform" "echo_two" {
        value = "${step.transform.echo.value}: Hello World"
    }

    output "val" {
        value = step.transform.echo.value
    }

    output "val_two" {
        value = step.transform.echo_two.value
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