pipeline "for_each_list" {

  step "echo" "for_each_echo" {
    for_each = ["a", "b", "c"]
    text = "${each.value}"
  }

  output "for_each_output" {
    value = step.echo.for_each_echo
  }

  output "for_each_output_1" {
    value = "${step.echo.for_each_echo[1].text}"
  }
}

pipeline "for_each_map" {

  step "echo" "for_each_echo" {
    for_each = {
        "a" = "b",
        "c" = "d"
    }
    text = "${each.key}: ${each.value}"
  }

  output "for_each_output" {
    value = step.echo.for_each_echo
  }

  output "for_each_output_1" {
    value = step.echo.for_each_echo["a"].text
  }
}

pipeline "parent_pipeline" {

  step "echo" "simple_echo" {
    text = "This is a simple echo step defined in parent pipeline"
  }

  step "sleep" "simple_sleep" {
    duration = "2s"
  }

  step "pipeline" "child_pipeline_itr_list" {
    pipeline = pipeline.for_each_list
  }

  step "echo" "echo_child_pipeline_1" {
    text = step.pipeline.child_pipeline_itr_list.for_each_output_1
  }

  step "pipeline" "child_pipeline_itr_map" {
    pipeline = pipeline.for_each_map
  }

  step "echo" "echo_child_pipeline_2" {
    text = step.pipeline.child_pipeline_itr_map.for_each_output_1
  }

  // step "pipeline" "foreach_pipeline_step"{
  //   pipeline = pipeline.for_each_map
  //   for_each = ["a", "b", "c"]
  // }

}
