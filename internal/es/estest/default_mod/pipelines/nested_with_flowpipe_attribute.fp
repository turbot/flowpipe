pipeline "parent_of_nested" {

    step "pipeline" "call_nested" {
        pipeline = pipeline.sleep_with_flowpipe_attributes
    }

    output "val_start" {
        value = step.pipeline.call_nested.flowpipe.started_at
    }

    output "val_end" {
        value = step.pipeline.call_nested.flowpipe.finished_at
    }
}
