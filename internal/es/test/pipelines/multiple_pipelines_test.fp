pipeline "parent_pipeline" {

    step "echo" "parent_echo" {
        text = "parent"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.child_pipeline
    }
}

pipeline "child_pipeline" {
    step "echo" "child_echo" {
        text = "child"
    }
}