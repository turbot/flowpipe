pipeline "redact" {

    step "transform" "guid" {
        value = "this should not be redacted: 9d9bdaa9-fa12-436b-bce8-9e783695b3ff"
    }

    output "val" {
        value = step.transform.guid.value
    }
}