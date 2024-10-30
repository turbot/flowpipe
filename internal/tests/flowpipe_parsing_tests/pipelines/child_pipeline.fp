pipeline "parent" {

  step "transform" "parent_echo" {
    value = "parent"
  }

  step "pipeline" "child_pipeline" {
    pipeline = pipeline.child
  }
}

pipeline "child" {
  step "transform" "child_echo" {
    value = "child"
  }
}

pipeline "child_step_with_args" {


  step "pipeline" "child_pipeline" {
    pipeline = pipeline.child

    args = {
      message = "this is a test"
      age     = 24
    }
  }
}
