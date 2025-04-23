mod "invalid_cred_types_dynamic" {
  title = "invalid_cred_types_dynamic"
}

pipeline "with_invalid_cred_type_dynamic" {

  param "cred" {
    type    = string
    default = "default"
  }

  step "transform" "test_creds" {
    value = credential.foo[param.cred].token
  }
}
