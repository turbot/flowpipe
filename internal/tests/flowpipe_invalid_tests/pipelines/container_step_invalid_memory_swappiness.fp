pipeline "pipeline_step_container" {

  description = "Container step test pipeline"

  step "container" "container_test1" {
    image             = "test/image"
    timeout           = 60000 // in ms
    memory_swappiness = 101
  }
}
