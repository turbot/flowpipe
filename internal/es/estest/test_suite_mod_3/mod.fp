mod "test_suite_mod_3" {
    title = "Test Suite Mod 3"

    require {
        mod "mod_depend_a" {
            version = "1.0.0"
            args = {
            }
        }
    }
}

pipeline "foo" {

    step "transform" "foo" {
        value = "foo"
    }

    output "val" {
        value = step.transform.foo.value
    }
}

pipeline "foo_calls_bar" {

    step "pipeline" "bar" {
        pipeline = mod_depend_a.pipeline.bar_a
    }

    output "val" {
        value = step.pipeline.bar.output.val
    }
}

pipeline "master" {
    param "execute" {
        type = string
        default = "skip"
    }
    step "pipeline" "respond" {
        pipeline = pipeline.respond
        args = {
            actions = {
                "skip" = {
                    pipeline_ref  = mod_depend_a.pipeline.optional_message_a
                    pipeline_args = {
                        notifier = "master"
                        send     = false
                        text     = "Skipped item - from top"
                    }
                }
                "enforce" = {
                    pipeline_ref  = mod_depend_a.pipeline.enforce_a
                    pipeline_args = {
                        notifier = "master"
                        send     = false
                        text     = "Enforced item."
                    }
                }
            }
            execute = param.execute
        }
    }

    output "val" {
        value = step.pipeline.respond.output.val
    }
}


pipeline "optional_message" {
    param "notifier" {
        type = string
        default = "foo"
    }

    param "send" {
        type = bool
        default = true
    }

    param "text" {
        type = string
        default = "Hello World"
    }

    step "transform" "output" {
        value = "${param.text} - ${param.notifier} - ${param.send}"
    }

    output "val" {
        value = step.transform.output.value
    }
}

pipeline "enforce" {
    param "notifier" {
        type = string
        default = "enforce"
    }

    param "send" {
        type = bool
        default = true
    }

    param "text" {
        type = string
        default = "Enforce Pipeline"
    }

    step "transform" "output" {
        value = "${param.text} - ${param.notifier} - ${param.send}"
    }

    output "val" {
        value = step.transform.output.value
    }
}

pipeline "respond" {
    param "execute" {
        type = string
        default = "skip"
    }

    param "actions" {
        description = "A map of actions, if approvers are set these will be offered as options to select, else the one matching the default_action will be used."
        type = map(object({
        pipeline_ref  = any
        pipeline_args = any
        }))
        default = {
        "skip" = {
            pipeline_ref  = pipeline.optional_message
            pipeline_args = {
            notifier = "default"
            send     = false
            text     = "Skipped item."
            }
        }
        }
    }

    step "pipeline" "action" {
        pipeline = param.actions[param.execute].pipeline_ref
        args = param.actions[param.execute].pipeline_args
    }

    output "val" {
        value = step.pipeline.action.output.val
    }
}