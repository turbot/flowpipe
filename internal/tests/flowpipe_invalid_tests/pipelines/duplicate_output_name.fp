pipeline "pipeline_with_duplicate_output_name" {

  step "transform" "print_hello" {
    value = "hello"
  }

  output "output_test" {
    value = step.transform.print_hello.value
  }

  output "output_test" {
    value = "hello world"
  }
}