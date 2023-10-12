pipeline "top" {

    step "echo" "hello" {
        text = "hello world"
    }

    step "pipeline" "middle" {
        pipeline = pipeline.middle
    }


    step "echo" "combine" {
        text = step.pipeline.middle.val
    }

    output "val" {
        value = step.echo.combine.text
    }
}

pipeline "middle" {

    step "echo" "echo" {
        text = "middle world"
    }

    output "val" {
        value = step.echo.echo.text
    }
}