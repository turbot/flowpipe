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

pipeline "child_step_with_args" {


    step "pipeline" "child_pipeline" {
        pipeline = pipeline.child

        args = {
            message = "this is a test"
            age = 24
        }
    }
}