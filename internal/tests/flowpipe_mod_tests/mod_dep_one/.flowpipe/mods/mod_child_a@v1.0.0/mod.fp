mod "mod_child_a" {
  title = "Child Mod A"
}

pipeline "this_pipeline_is_in_the_child" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_a" {
        value = step.transform.foo.value
    }
}
