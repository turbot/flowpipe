pipeline "notifier_param" {

    param "notifier" {
        type = notifier
        default = notifier.frontend
    }

    step "transform" "notifier" {
        value = param.notifier.title
    }

    output "val" {
        value = step.transform.notifier
    }
}

pipeline "notifier_list_param" {

    param "notifiers" {
        type = list(notifier)
        default = [notifier.frontend]
    }

    step "transform" "notifiers" {
        value = param.notifiers
    }

    output "val" {
        value = step.transform.notifiers
    }
}