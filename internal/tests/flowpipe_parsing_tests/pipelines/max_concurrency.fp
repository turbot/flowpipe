pipeline "step_with_max_concurrency" {

  step "container" "step_1" {
    for_each        = range(0, 100)
    image           = "alpine:3.12"
    max_concurrency = 15
  }

  step "container" "step_2" {
    for_each        = range(0, 100)
    image           = "alpine:3.12"
  }
}

pipeline "pipeline_with_max_concurrency" {

  max_concurrency = 15

  step "container" "step_1" {
    for_each        = range(0, 100)
    image           = "alpine:3.12"
    max_concurrency = 15
  }

  step "container" "step_2" {
    for_each        = range(0, 100)
    image           = "alpine:3.12"
  }
}

