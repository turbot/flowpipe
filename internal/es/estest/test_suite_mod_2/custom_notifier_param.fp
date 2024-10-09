variable "notifier" {
    type = notifier
}

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

pipeline "notifier_param_parent" {

    param "notifier" {
        type = notifier
        default = notifier.default
    }

    step "pipeline" "call_child" {
        pipeline = pipeline.notifier_param_child
        args = {
            child_notifier = param.notifier
        }
    }

   output "val" {
        value = step.pipeline.call_child.output.val
    }
}
pipeline "notifier_param_child" {

    param "child_notifier" {
        type = notifier
    }

    step "transform" "t" {
        value = param.child_notifier
    }

    output "val" {
        value = step.transform.t
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


pipeline "notifier_var_param" {

    param "notifier" {
        type = notifier
        default = var.notifier
    }

    step "transform" "notifier" {
        value = param.notifier.title
    }

    output "val" {
        value = step.transform.notifier
    }
}

pipeline "notifier_var" {

    step "transform" "notifier" {
        value = var.notifier.title
    }

    output "val" {
        value = step.transform.notifier
    }
}