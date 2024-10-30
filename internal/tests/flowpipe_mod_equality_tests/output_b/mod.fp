mod "test" {

}

pipeline "pipeline_loop_test" {

  param "message" {
    type = string
  }

  param "index" {
    type = number
  }

  output "greet_world" {
    value =  "Hello world! ${param.message} ${param.index}"
  }
}
