pipeline "empty_slice" {

    step "transform" "empty_list" {
        value = []
    }

    output "val" {
        value = step.transform.empty_list.value
    }
}