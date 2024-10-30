mod "dynamic_pipe_ref" {

}

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

    output "val" {
        value = step.pipeline.middle_dynamic_static_to_a
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