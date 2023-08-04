pipeline "parent_pipeline" {

    step "echo" "parent_echo" {
        text = "parent echo step"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.child_pipeline
    }


    output "parent_output" {
        value = step.pipeline.child_pipeline.child_output
    }
}

pipeline "child_pipeline" {
    step "echo" "child_echo" {
        text = "child echo step"
    }

    output "child_output" {
        value = step.echo.child_echo.text
    }
}