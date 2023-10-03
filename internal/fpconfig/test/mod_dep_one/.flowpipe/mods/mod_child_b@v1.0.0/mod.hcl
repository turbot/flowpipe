mod "mod_child_b" {
  title = "Child Mod B"
}

pipeline "this_pipeline_is_in_the_child_b" {
    step "echo" "foo" {
        text = "foo"
    }

    step "echo" "baz" {
        text = "baz"
    }

    output "foo_a" {
        value = step.echo.foo.text
    }

    # this will test resolving a step in a child pipeline
    # when the mod is a dependency of the parent
    step "pipeline" "child_pipeline" {
        pipeline = pipeline.foo_two
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
