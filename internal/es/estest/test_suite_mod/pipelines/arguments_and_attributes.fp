pipeline "refer_to_arguments" {

    step "http" "get" {
        url = "http://api.open-notify.org/astros.json"
    }

    step "transform" "refer_to_arguments" {
        value = step.http.get.url
    }

    output "val" {
        value = step.transform.refer_to_arguments.value
    }
}