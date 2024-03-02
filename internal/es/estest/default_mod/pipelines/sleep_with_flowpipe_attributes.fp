pipeline "sleep_with_flowpipe_attributes" {

    step "sleep" "sleep" {
        duration = "1s"
    }

    step "transform" "check_start" {
        value = step.sleep.sleep.flowpipe.started_at
    }

    step "transform" "check_finish" {
        value = step.sleep.sleep.flowpipe.finished_at
    }

    output "val_start" {
        value = step.transform.check_start.value
    }

    output "val_end" {
        value = step.transform.check_finish.value
    }
}