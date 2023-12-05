pipeline "simple_error" {
    step "http" "does_not_exist" {
        url = "https://google.com/bad.json"
    }

    output "val" {
        value = "should not be calculated"
    }
}
