trigger "schedule" "simple" {
  pipeline = pipeline.simple_with_trigger
  args = {
      param_one = "from trigger"
  }
}
