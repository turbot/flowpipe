
pipeline "in_b" {

    step "transform" "test_b" {
        value = "echo b v1.0.0"
    }

    output "val" {
        value = step.transform.test_b
    }
}

pipeline "in_b_with_pipe_as_param" {

    param "action" {
         type = object({
            label         = string
            pipeline_ref  = any
        })
        default = {
            label = "echo"
            pipeline_ref = pipeline.in_b
        }
    }

    step "pipeline" "pipe" {
        pipeline = param.action.pipeline_ref
    }

    output "val" {
        value = step.pipeline.pipe.output.val.value
    }
}