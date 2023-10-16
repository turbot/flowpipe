pipeline "input_one" {

    step "echo" "foo" {
        text = "foo"
    }

    step "input" "input" {
        type = button
        destination = slack
        
    }
}

