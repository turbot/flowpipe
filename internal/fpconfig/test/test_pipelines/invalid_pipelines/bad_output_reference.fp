pipeline "bad_output_reference" {


    step "echo" "echo" {
        text = "Hello World"
    }

    output "echo" {
        value = step.echo.does_not_exist
    }
}