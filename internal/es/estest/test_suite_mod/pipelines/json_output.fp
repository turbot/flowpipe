pipeline "json_output" {

    step "transform" "source" {
        value = "{\n  \"name\": \"bob\",\n  \"age\": 50,\n  \"address\": {\n    \"country\": \"uk\",\n    \"city\": \"London\"\n  }\n}"
    }

    step "transform" "json" {
        value = ""
        output "val" {
            value = jsondecode(step.transform.source.value)
        }
    }

    output "country" {
        value = step.transform.json.output.val.address.country
    }

    output "all" {
        value = jsondecode(step.transform.source.value)
    }
}

pipeline "parent_json_output" {

    step "pipeline" "call_json_output" {
        pipeline = pipeline.json_output
    }

    output "city" {
        value = step.pipeline.call_json_output.output.all.address.city
    }
}
