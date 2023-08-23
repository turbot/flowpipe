mod "mod_one" {
  title = "Mod One"
}


pipeline "astros" {
    description = "Astro Pipeline"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.json"
    }

    output "child_output" {
        value = step.http.my_step_1.status_code
    }
}
