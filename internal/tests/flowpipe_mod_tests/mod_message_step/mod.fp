mod "mod_message_step" {

}

pipeline "message_step_one" {

    step "message" "hello" {
        notifier = notifier.default
        text = "Hello World"
    }
    
    output "val" {
        value = "Hello World!"
    }
}


pipeline "message_step_with_overrides" {

    step "message" "hello" {
        notifier = notifier.default
        text = "Hello World 2"

        cc = ["foo", "baz"]
        bcc = ["bar"]

        channel = "channel override"
    }
    
    output "val" {
        value = "Hello World!"
    }
}

pipeline "message_step_with_throw" {

    step "message" "hello" {
        notifier = notifier.default
        text = "Hello World"

        throw {
            if      = result.text == "Hello World"
            message = "Message is Hello World"
        }
    }
    
    output "val" {
        value = "Hello World!"
    }
}


pipeline "message_step_with_error" {

    step "message" "hello" {
        notifier = notifier.default
        text = "Hello World"

        error {
            ignore = true
        }
    }
    
    output "val" {
        value = "Hello World!"
    }
}
