mod "mod_with_creds" {
  title = "mod_with_creds"
}


pipeline "with_creds" {

    step "transform" "echo" {
        value = credential.aws.abc.access_key
    }
}


