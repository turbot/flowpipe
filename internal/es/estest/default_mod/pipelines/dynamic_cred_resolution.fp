pipeline  "dynamic_cred_parent" {

    step "pipeline" "call_dynamic_cred" {
        pipeline = pipeline.dynamic_cred

        // simulate acquiring multiple connections from a query step
        for_each = ["sso", "dundermifflin"]

        args = {
            creds = each.value
        }
    }

    output "val_0" {
        value = step.pipeline.call_dynamic_cred[0].output.val
    }

    output "val_1" {
        value = step.pipeline.call_dynamic_cred[1].output.val
    }
}

pipeline "dynamic_cred" {

    param "creds" {
        type = string
        default = "example"
    }

    step "transform" "test" {
        output "val" {
            value = credential.aws[param.creds]
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }
}
