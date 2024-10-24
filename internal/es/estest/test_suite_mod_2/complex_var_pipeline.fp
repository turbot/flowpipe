pipeline "complex_var" {

    step "transform" "val" {
        value = var.complex_var
    }

    output "val" {
        value = step.transform.val
    }
}