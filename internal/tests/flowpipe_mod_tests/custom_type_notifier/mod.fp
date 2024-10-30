mod "custom_type_notifier" {

}

pipeline "notifier" {
    param "notifier" {
        type    = notifier
        default = notifier.default
    }

    param "list_of_notifiers" {
        type    = list(notifier)
        default = [
            notifier.default,
        ]
    }

    param "list_of_notifiers_more" {
        type    = list(notifier)
        default = [
            notifier.default,
            notifier.admin,
        ]
    }
}

