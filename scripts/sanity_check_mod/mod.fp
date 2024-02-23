mod "sanity_check" {

}

pipeline "cred" {

    step "transform" "test" {
        output "val" {
            value = credential.aws.example
        }
    }

    // step "transform" "cred" {
    //     value = credential.aws.example
    // }

    // output "val" {
    //     value = step.transform.cred.value
    // }
}