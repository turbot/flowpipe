pipeline "short_circuit" {
    step "transform" "foo" {
        value = true
    }

    step "transform" "bar" {
        if = false
        value = true
        error {
            ignore = true
        }
    }

    output "changed" {
        value = anytrue([
            step.transform.foo.value,
            step.transform.bar.value
        ])
    }
}