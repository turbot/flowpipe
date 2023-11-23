pipeline "for_each_with_sleep" {

    step "sleep" "sleep" {
        for_each = ["1s", "2s", "3s"]
        duration = each.value
    }

    step "transform" "echo" {
        depends_on = [step.sleep.sleep]
        value      = "ends"
    }

    output "val_sleep" {
        value = step.sleep.sleep
    }

    output "val" {
        value = step.transform.echo
    }

}