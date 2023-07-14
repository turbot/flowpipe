pipeline "for_loop" {

    param "users" {
        type    = "list"
        default = ["jerry","Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = param.users
        text = "user if ${each.value}"
    }
}