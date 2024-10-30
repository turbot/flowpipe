pipeline "bad_output_reference" {
    step "transform" "echo" {
        value = "Hello World"
    }

    output "echo" {
        value = step.transform.does_not_exist
    }
}