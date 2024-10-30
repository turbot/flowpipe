mod "test_mod" {
}

pipeline "conn_in_step_output" {

    step "transform" "test" {
        output "val" {
            value = connection.aws.example
        }
    }

    output "val" {
        value = step.transform.test.val
    }
}

pipeline "conn_in_output" {
    output "val" {
        value = connection.aws.example
    }
}