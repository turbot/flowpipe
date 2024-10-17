
mod "mod_depend_x" {

}

pipeline "display_x" {

    step "transform" "echo_x" {
        value = "echo from x v1.0.0"
    }

    output "val" {
        value = step.transform.echo_x
    }
}