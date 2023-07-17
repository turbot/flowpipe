pipeline "two_text" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = "baz ${step.echo.text_1.text}"
    }
}

pipeline "three_text" {
    step "echo" "text_1" {
        text = "foo"
    }

    step "echo" "text_2" {
        text = "baz ${step.echo.text_1.text}"
    }

    step "echo" "text_3" {
        text = "text_1: ${step.echo.text_1.text} text_2: ${step.echo.text_2.text}"
    }
}

pipeline "http_depends" {
    step "echo" "text_1" {
        text = "astros.json"
    }

    step "http" "http_1" {
        url = "http://api.open-notify.org/${step.echo.text_1.text}"
    }
}

pipeline "sleep_depends" {
    step "echo" "text_1" {
        text = "1s"
    }

    step "sleep" "sleep_1" {
        duration = step.echo.text_1.text
    }
}



pipeline "http_and_sleep" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        duration = "2s"
    }
}

pipeline "http_and_sleep_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [step.http.http_1]
        duration = "2s"
    }
}

pipeline "http_and_sleep_multiple_depends" {
    description = "http and sleep pipeline"
    step "http" "http_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "sleep" "sleep_1" {
        depends_on = [step.http.http_1]
        duration = "2s"
    }

    step "http" "http_2" {
        url = "http://api.open-notify.org/astros.json"
        depends_on = [
            step.http.http_1,
            step.sleep.sleep_1
        ]
    }

    step "http" "http_3" {
        url = "http://api.open-notify.org/astros.json"
        depends_on = [
            step.http.http_1,
            step.sleep.sleep_1,
            step.http.http_2
        ]
    }
}


pipeline "two_sleeps" {
    step "sleep" "sleep_1" {
        duration = "1s"
    }

    step "sleep" "sleep_2" {
        duration = "1s"
    }
}