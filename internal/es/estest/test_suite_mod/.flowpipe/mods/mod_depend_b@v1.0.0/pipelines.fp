
pipeline "in_b" {

    step "transform" "test_b" {
        value = "echo b"
    }

    output "val" {
        value = step.transform.test_b
    }
}