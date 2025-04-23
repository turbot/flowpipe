mod "mod_child_b" {
  title = "Child Mod B"
}

pipeline "this_pipeline_is_in_the_child_b" {
    step "transform" "foo" {
        value = "foo"
    }

    step "transform" "baz" {
        value = "baz"
    }

    output "foo_a" {
        value = step.transform.foo.value
    }

    # this will test resolving a step in a child pipeline
    # when the mod is a dependency of the parent
    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two
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
