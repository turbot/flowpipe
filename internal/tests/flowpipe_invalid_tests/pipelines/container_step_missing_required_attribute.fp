pipeline "container_step_missing_required_attribute" {

  description = "Container step with missing source/image"

  step "container" "source_test" {
    cmd = []
  }
}
