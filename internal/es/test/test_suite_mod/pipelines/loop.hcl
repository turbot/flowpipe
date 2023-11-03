pipeline "simple_loop" {

    step "echo" "repeat" {
        text  = "iteration: ${loop.index}"
        numeric = 1

        loop {
            if = result.numeric < 3
            numeric = result.numeric + 1
        }
    }

    output "val_1" {
        value = step.echo.repeat["0"]
    }
    output "val_2" {
        value = step.echo.repeat["1"]
    }
    output "val_3" {
        value = step.echo.repeat["2"]
    }
}

pipeline "simple_loop_index" {

    step "echo" "repeat" {
        text  = "iteration: ${loop.index}"

        loop {
            if = loop.index < 2
        }
    }

    output "val_1" {
        value = step.echo.repeat["0"]
    }
    output "val_2" {
        value = step.echo.repeat["1"]
    }
    output "val_3" {
        value = step.echo.repeat["2"]
    }
}


pipeline "loop_with_for_each" {

    step "echo" "repeat" {
        for_each = ["oasis", "blur"]
        text = "iteration: ${loop.index} - ${each.value}"

        loop {
            if = loop.index < 2
        }
    }

    output "val" {
        value = step.echo.repeat
    }
}

pipeline "loop_with_for_each_sleep" {

    step "sleep" "repeat" {
        for_each = ["1s", "4s"]
        duration = each.value

        loop {
            if = loop.index < 2
        }
    }

    output "val" {
        value = step.sleep.repeat
    }
}