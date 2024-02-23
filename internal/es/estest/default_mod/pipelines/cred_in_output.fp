pipeline "cred_in_step_output" {

    step "transform" "test" {
        output "val" {
            value = credential.aws.example
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }

    // step "transform" "cred" {
    //     value = credential.aws.example
    // }

    // output "val" {
    //     value = step.transform.cred.value
    // }
}