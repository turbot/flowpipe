mod "test_mod" {
}

pipeline "cred_in_step_output" {

    step "transform" "test" {
        output "val" {
            value = credential.aws.example
        }
    }

    output "val" {
        value = step.transform.test.val
    }
}

pipeline "cred_in_output" {
    output "val" {
        value = credential.aws.example
    }
}