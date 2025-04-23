pipeline "empty_slice" {

    param "empty_input_string" {
        type = list(string)
        default = []
    }

    param "empty_input_number" {
        type = list(number)
        default = []
    }

    step "transform" "empty_list" {
        value = []
    }

    step "transform" "empty_list_string" {
        value = param.empty_input_string
    }    

    output "val" {
        value = step.transform.empty_list.value
    }

    output "empty_output" {
        value = []
    }
}
