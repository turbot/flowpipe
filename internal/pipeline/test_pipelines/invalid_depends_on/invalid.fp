pipeline "invalid_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [
            step.http.my_step_1
        ]
        duration = 2
    }
}
