pipeline "json_output" {

    step "echo" "source" {
        text = "{\n  \"name\": \"bob\",\n  \"age\": 50,\n  \"address\": {\n    \"country\": \"uk\",\n    \"city\": \"London\"\n  }\n}"
    }

    step "echo" "json" {
        text = ""
        output "val" {
            value = jsondecode(step.echo.source.text)
        }
    }

    output "country" {
        value = step.echo.json.output.val.address.country
    }

    output "all" {
        value = jsondecode(step.echo.source.text)
    }
}

pipeline "parent_json_output" {

    step "pipeline" "call_json_output" {
        pipeline = pipeline.json_output
    }

    output "city" {
        value = step.pipeline.call_json_output.all.address.city
    }
}
