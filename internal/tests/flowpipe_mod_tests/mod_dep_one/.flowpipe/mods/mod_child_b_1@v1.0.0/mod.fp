mod "mod_child_b" {
  title = "Child Mod B"
}

pipeline "this_pipeline_is_in_the_child" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_a" {
        value = step.transform.foo.value
    }
}
