

mod "pipeline_with_references" {
    title = "Test mod"
    description = "Use this mod for testing references within pipeline and from one pipeline to another"
}


pipeline "foo" {

    step "transform" "bar" {
        value = "test"
    }


    step "transform" "baz" {
        value = step.transform.barrs
    }
}