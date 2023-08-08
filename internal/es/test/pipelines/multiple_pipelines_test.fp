pipeline "parent_pipeline" {
    description = "Parent pipeline with a child pipeline"
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
    description = "Child Pipeline"
    step "echo" "child_echo" {
        text = "child echo step"
    }

    output "child_output" {
        value = step.echo.child_echo.text
    }
}


pipeline "parent_multiple_pipelines_with_errors" {
    description = "Parent pipeline with multiple child pipelines with errors"
    step "echo" "parent_echo" {
        text = "parent echo step"
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
    step "echo" "child_echo" {
        text = "child A echo step"
    }

    output "child_output" {
        value = step.echo.child_echo.text
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
