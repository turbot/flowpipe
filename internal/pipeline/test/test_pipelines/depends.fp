pipeline "implicit_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [
            step.http.http_1
        ]
        duration = "2s"
    }

    step "sleep" "sleep_2" {
        duration = step.sleep.sleep_1.duration
    }
}

pipeline "depends_index" {
    param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "echo" "echo_1" {
        text = "sleep 1 output: ${step.sleep.sleep_1[0].duration}"
    }
}

pipeline "explicit_depends_index" {
    param "time" {
        type = list(string)
        default = ["1s", "2s"]
    }

    step "sleep" "sleep_1" {
        for_each = param.time
        duration = each.value
    }

    step "echo" "echo_1" {
        depends_on = [
            step.sleep.sleep_1[0]
        ]
        text = "sleep 1 foo"
    }
}