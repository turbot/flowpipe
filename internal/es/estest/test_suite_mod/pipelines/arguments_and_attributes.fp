pipeline "refer_to_arguments" {

    step "http" "get" {
        url = "http://localhost:7104/check.json"
    }

    step "transform" "refer_to_arguments" {
        value = step.http.get.url
    }

    output "val" {
        value = step.transform.refer_to_arguments.value
    }
}