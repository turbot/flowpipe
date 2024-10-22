pipeline "a_calls_b" {
    step "transform" "test" {
        value = "echo"
    }

    step "pipeline" "in_b" {
        pipeline = mod_depend_b.pipeline.in_b
    }

    output "out" {
        value = step.pipeline.in_b
    }
}

pipeline "a_calls_b_pass_x" {

    step "pipeline" "in_b_pass_x" {
        pipeline = mod_depend_b.pipeline.in_b_with_pipe_as_param

        args = {
            action = {
                label = "echo",
                pipeline_ref = mod_depend_x.pipeline.display_x
            }
        }
    }

    output "out" {
        value = step.pipeline.in_b_pass_x.output.val
    }
}