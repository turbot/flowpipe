mod "test" {

}

pipeline "bad_step_ref" {
  step "transform" "one" {
    value = "one"
  }

  step "transform" "two" {
    value = step.transform.onex.value
  }
}