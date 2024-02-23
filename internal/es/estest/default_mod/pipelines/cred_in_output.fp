pipeline "cred_in_step_output" {

    step "transform" "test" {
        output "val" {
            value = credential.aws.example
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }
}

pipeline "cred_in_output" {
    output "val" {
        value = credential.aws.example.access_key
    }
}