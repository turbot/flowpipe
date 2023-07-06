pipeline "implicit_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [
            step.http.http_1
        ]
        duration = 2
    }

    step "sleep" "sleep_2" {
        duration = step.sleep.sleep_1.duration
    }
}


pipeline "implicit_depends_text" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [
            step.http.http_1
        ]
        duration = 2
    }

    step "sleep" "sleep_2" {
        duration = step.sleep.sleep_1.duration
    }
}

