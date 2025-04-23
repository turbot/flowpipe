

mod "pipeline_with_references" {
  title       = "Test mod"
  description = "Use this mod for testing references within pipeline and from one pipeline to another"
}


pipeline "foo" {

  step "transform" "baz" {
    value = step.transform.bar
  }

  step "transform" "bar" {
    value = "test"
  }

  step "pipeline" "child_pipeline" {
    pipeline = pipeline.foo_two_invalid
  }

  step "transform" "child_pipeline" {
    value = step.pipeline.child_pipeline.foo
  }
}


pipeline "foo_two" {
  step "transform" "baz" {
    value = "foo"
  }

  output "foo" {
    value = transform.baz.value
  }
}
