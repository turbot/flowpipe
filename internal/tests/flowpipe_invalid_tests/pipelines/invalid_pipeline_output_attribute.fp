pipeline "invalid_pipeline_output_attribute_test" {

  step "transform" "print_hello" {
    value = "hello"
  }

  output "invalid_output" {
    depends_on = step.transform.print_hello.value
    value      = "hello world"
  }
}