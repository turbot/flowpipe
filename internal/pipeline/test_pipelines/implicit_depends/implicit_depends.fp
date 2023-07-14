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








pipeline "implicit_depends_text" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "echo" "my_text" {
        text = "5"
    }

    step "sleep" "sleep_1" {
        description = "bar ${step.echo.my_text.output} baz"
        duration = foo("${step.echo.my_text.output}m")
    }

    step "baz" "my_baz" {
        input = step.sleep.sleep_1.description
    }
}

