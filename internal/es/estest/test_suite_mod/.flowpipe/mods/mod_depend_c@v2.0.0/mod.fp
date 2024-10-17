mod "mod_depend_c" {

    require {
        mod "mod_depend_b" {
            version = "2.0.0"
        }
    }
}

pipeline "c_calls_b" {
    step "transform" "test" {
        value = "echo"
    }

    step "pipeline" "in_b" {
        pipeline = mod_depend_b.pipeline.in_b
    }

    output "out" {
        value = step.pipeline.in_b
    }
}
