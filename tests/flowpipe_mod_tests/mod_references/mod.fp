

mod "pipeline_with_references" {
    title = "Test mod"
    description = "Use this mod for testing references within pipeline and from one pipeline to another"
}


pipeline "foo" {

    # leave this here to ensure that references that is later than the resource can be resolved
    #
    # we parse the HCL files from top to bottom, so putting this step `baz` after `bar` is the easier path
    # reversing is the a harder parse
    step "transform" "baz" {
        value = step.transform.bar
    }

    step "transform" "bar" {
        value = "test"
    }

    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two
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
