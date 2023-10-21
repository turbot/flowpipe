

mod "pipeline_with_references" {
    title = "Test mod"
    description = "Use this mod for testing references within pipeline and from one pipeline to another"
}


pipeline "foo" {

    step "echo" "baz" {
        text = step.echo.bar
    }

    step "echo" "bar" {
        text = "test"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two_invalid
    }

    step "echo" "child_pipeline" {
        text = step.pipeline.child_pipeline.foo
    }
}


pipeline "foo_two" {
    step "echo" "baz" {
        text = "foo"
    }

    output "foo" {
        value = echo.baz.text
    }
}
