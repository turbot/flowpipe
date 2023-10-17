pipeline "input_one" {
    step "echo" "echo" {
        text = "hello"
    }

    step "input" "input" {
        // type = button
        // destination = slack

    }
}

