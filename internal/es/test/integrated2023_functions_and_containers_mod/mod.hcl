mod "demo" {
    title = "Mod for Integrated 2023 Flowpipe - Functions and Containers demo."
}

pipeline "foo" {
    step "echo" "echo" {
        text = "bar"
    }
}