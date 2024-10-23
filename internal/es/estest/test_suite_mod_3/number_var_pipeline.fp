pipeline "number_var_pipeline" {

    param "number_var" {
        type = number
        description = "A number variable."
        default = var.number_var
    }

    output "number_var_output" {
        value = var.number_var
    }
}
