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
    value = param.index <= 5 ?  "Hello world! ${param.message} ${param.index}" : null
  }
}
