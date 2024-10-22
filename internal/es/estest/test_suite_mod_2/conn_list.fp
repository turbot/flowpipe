pipeline "conn_list" {


    step "transform" "list" {
        value = connection.aws
    }

    output "val" {
        value = step.transform.list.value
    }

}
