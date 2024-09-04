pipeline "conn_in_step_output" {

    step "transform" "test" {
        output "val" {
            value = connection.aws.example
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }
}

pipeline "conn_in_output" {
    output "val" {
        value = connection.aws.example.access_key
    }
}