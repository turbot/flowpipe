pipeline "if" {
    param "condition_true" {
        type = bool
        default = true
    }

    step "transform" "text_true" {
        value = "foo"
        if    = param.condition_true
    }

    param "condition_false" {
        type = bool
        default = false
    }

    step "transform" "text_false" {
        value = "foo"
        if    = param.condition_false
    }

    step "transform" "text_1" {
        value = "foo"
    }

    step "transform" "text_2" {
        value = "bar"
        if    = step.transform.text_1.value == "foo"
    }

    step "transform" "text_3" {
        value = "baz"
        if    = step.transform.text_1.value == "bar"
    }
}

trigger "schedule" "every_hour_trigger_on_if" {
    description = "trigger that will run every hour"
    schedule    = "hourly"
    pipeline    = pipeline.if
}
