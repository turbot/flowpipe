mod "test" {

}

pipeline "bad_step_ref" {

  step "transform" "output" { 

    value = "bar"

    output "foo" {
      value = step.transform.does_not_exist
    }
  }
  
}