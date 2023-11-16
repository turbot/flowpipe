pipeline "simple_loop" {

    step "echo" "repeat" {
        text  = "iteration: ${loop.index}"
        numeric = 1

        loop {
            until = result.numeric < 3
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
            until = loop.index < 2
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
        for_each = ["oasis", "blur", "radiohead"]
        text = "iteration: ${loop.index} - ${each.value}"

        loop {
            until = loop.index < 3
        }
    }

    output "val" {
        value = step.echo.repeat
    }
}

pipeline "lots_of_for_each" {

    step "echo" "repeat" {
        for_each = ["oasis", "blur", "radiohead", "the verve", "the beatles", "the rolling stones", "the sex pistols"]
        text = "name: ${each.value}"
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
            until = loop.index < 2
        }
    }

    output "val" {
        value = step.sleep.repeat
    }
}

pipeline "loop_with_for_each_and_nested_pipeline" {

    step "pipeline" "repeat" {
        for_each = ["oasis", "blur", "radiohead"]
        # for_each = ["oasis"]
        pipeline = pipeline.nested_echo

        args = {
            name = each.value
            loop_index = loop.index
        }

        loop {
            until = loop.index < 2
        }
    }

    output "val" {
        value = step.pipeline.repeat
    }
}

pipeline "nested_echo" {

    param "name" {
        type = string
    }

    param "loop_index" {
        type = number
    }

    step "transform" "echo" {
        value = "${param.loop_index}: ${param.name}"
    }

    output "val" {
        value = step.transform.echo.value
    }
}