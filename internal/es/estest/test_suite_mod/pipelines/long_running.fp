pipeline "long_sleep" {
    step "sleep" "long" {
        duration = "20s"
    }
    output "val" {
        value = "done"
    }
}
