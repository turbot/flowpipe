pipeline "a_calls_b" {
    step "transform" "test" {
        value = "echo"
    }

    step "pipeline" "in_b" {
        pipeline = mod_depend_b.pipeline.in_b
    }

    output "out" {
        value = step.transform.test
    }
}
