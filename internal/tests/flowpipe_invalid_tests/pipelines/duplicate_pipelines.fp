pipeline "pipeline_007" {
    step "transform" "name" {
        value = "I am Pipe, FlowPipe - the original"
    }

    output "val" {
        value = step.transform.name.value
    }
}

pipeline "pipeline_007" {
    step "transform" "name" {
        value = "I am Pipe, FlowPipe  - the dummy"
    }

    output "val" {
        value = step.transform.name.value
    }
}