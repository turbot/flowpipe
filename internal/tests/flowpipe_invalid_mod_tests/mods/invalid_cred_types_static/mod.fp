mod "invalid_cred_types_static" {
  title = "invalid_cred_types_static"
}

pipeline "with_invalid_cred_type_static" {

  step "transform" "test_creds" {
    value = credential.foo["default"].token
  }
}
