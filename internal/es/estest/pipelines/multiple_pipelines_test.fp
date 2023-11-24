pipeline "parent_pipeline" {
    description = "Parent pipeline with a child pipeline"
    step "transform" "parent_echo" {
        value = "parent echo step"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.child_pipeline
    }


    output "parent_output" {
        value = step.pipeline.child_pipeline.output.child_output
    }
}

pipeline "child_pipeline" {
    description = "Child Pipeline"
    step "transform" "child_echo" {
        value = "child echo step"
    }

    output "child_output" {
        value = step.transform.child_echo.value
    }
}


pipeline "parent_multiple_pipelines_with_errors" {
    description = "Parent pipeline with multiple child pipelines with errors"
    step "transform" "parent_echo" {
        value = "parent echo step"
    }

    step "pipeline" "child_pipeline_a" {
        pipeline = pipeline.child_pipeline_a
    }

    step "pipeline" "child_pipeline_b" {
        pipeline = pipeline.child_pipeline_b
    }

    step "pipeline" "child_pipeline_c" {
        pipeline = pipeline.child_pipeline_c
    }

}

pipeline "child_pipeline_a" {
    description = "Child Pipeline A"
    step "transform" "child_echo" {
        value = "child A echo step"
    }

    output "child_output" {
        value = step.transform.child_echo.value
    }
}

pipeline "child_pipeline_b" {
    description = "Child Pipeline B"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
    }

    output "child_output" {
        value = step.http.my_step_1.status_code
    }
}

pipeline "child_pipeline_c" {
    description = "Child Pipeline C"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    output "child_output" {
        value = step.http.my_step_1.status_code
    }
}


pipeline "parent_pipeline_with_args" {
    description = "Parent pipeline, calls a child pipeline with args"
    step "pipeline" "child_pipeline_with_args" {
        pipeline = pipeline.child_pipeline_with_args

        args = {
            message = "from parent"
            age = 24
        }
    }

    output "parent_output" {
        value = step.pipeline.child_pipeline_with_args.output.child_output
    }
}

pipeline "parent_pipeline_with_no_args" {
    description = "Parent pipeline, calls a child pipeline with args"

    step "pipeline" "child_pipeline_with_args" {
        pipeline = pipeline.child_pipeline_with_args
    }
}


pipeline "child_pipeline_with_args" {
    description = "Child Pipeline with Args"

    param "message" {
        type = string
        default = "change this message"
    }

    param "age" {
        type = number
        default = 1
    }


    step "transform" "child_echo" {
        value = "child echo step: ${param.message} ${param.age}"
    }

    output "child_output" {
        value = step.transform.child_echo.value
    }
}
