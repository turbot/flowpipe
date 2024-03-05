pipeline "for_each_with_null" {

    step "transform" "echo" {
        for_each = ["a", null, "c"]

        if = each.value != null

        value = each.value
    }

    output "val" {
        value = step.transform.echo
    }
}

pipeline "for_each_is_null" {

    param "for_each" {
        default = null
    }

    step "transform" "echo" {
        for_each = param.for_each == null ? [] : param.for_each

        value = each.value
    }

    output "val" {
        value = step.transform.echo
    }
}