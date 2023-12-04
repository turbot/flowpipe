pipeline "parent_pipeline_param_mismatch" {

    step "pipeline" "child_pipeline_param_mismatch" {
        pipeline = pipeline.child_pipeline_param_mismatch
        args = {
            invalid_param = "billy"
        }
    }

    output "val" {
        value = step.pipeline.child_pipeline_param_mismatch.output.val
    }
}

pipeline "child_pipeline_param_mismatch" {

    param "name" {
        type = string
    }

    step "transform" "echo" {
        value = "Hello ${param.name}"
    }

    output "val" {
        value = step.transform.echo.value
    }
}