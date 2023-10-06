pipeline "pipeline_007" {
    step "echo" "name" {
        text = "I am Pipe, FlowPipe - the original"
    }

    output "val" {
        value = step.echo.name.text
    }
}

pipeline "pipeline_007" {
    step "echo" "name" {
        text = "I am Pipe, FlowPipe  - the dummy"
    }

    output "val" {
        value = step.echo.name.text
    }
}