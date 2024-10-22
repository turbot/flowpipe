
pipeline "in_b" {

    step "transform" "test_b" {
        value = "echo b v2.0.0"
    }

    output "val" {
        value = step.transform.test_b
    }
}