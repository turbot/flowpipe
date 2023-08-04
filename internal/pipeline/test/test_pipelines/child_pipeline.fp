pipeline "parent" {

    step "echo" "parent_echo" {
        text = "parent"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.child
    }
}

pipeline "child" {
    step "echo" "child_echo" {
        text = "child"
    }
}