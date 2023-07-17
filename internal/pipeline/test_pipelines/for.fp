pipeline "for_loop" {

    var "users" {
        type = list(string)
        default = ["jerry","Janis", "Jimi"]
    }

    step "echo" "text_1" {
        for_each = var.users
        text = "user if ${each.value}"
    }

    step "echo" "no_for_each" {
        text = "baz"
    }
}