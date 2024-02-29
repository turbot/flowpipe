pipeline "top_dynamic" {

    param "pipe" {
        default = "middle_dynamic_b"
    }

    step "pipeline" "middle_dynamic_static_to_a" {
        pipeline = pipeline.middle_dynamic_a
    }

    step "pipeline" "middle_dynamic" {
        pipeline = pipeline[param.pipe]
    }

    output "val_a" {
        value = step.pipeline.middle_dynamic_static_to_a
    }

    output "val_b" {
        value = step.pipeline.middle_dynamic
    }
}

pipeline "top_dynamic_step_ref" {


    step "transform" "pipe_name" {
        value = "middle_dynamic_c"
    }

    step "pipeline" "middle_dynamic" {
        # pipeline name from a differen step
        pipeline = pipeline[step.transform.pipe_name.value]
    }

    output "val" {
        value = step.pipeline.middle_dynamic
    }
}


pipeline "middle_dynamic_a" {

    output "val" {
        value = "A"
    }
}

pipeline "middle_dynamic_b" {
    output "val" {
        value = "B"
    }
}

pipeline "middle_dynamic_c" {
    output "val" {
        value = "C"
    }
}