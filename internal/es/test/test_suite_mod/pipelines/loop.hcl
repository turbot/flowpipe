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
        value = step.echo.repeat["3"]
    }
}
