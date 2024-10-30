mod "test" {

}

pipeline "input_with_throw" {

    step "input" "my_input" {
        prompt   = "Shall we play a game?"
        type     = "select"
        notifier = notifier.default

        option "Tic Tac Toe" {}
        option "Checkers" {}
        option "Global Thermonuclear War" {}

        throw {
            if      = result.value == "Checkers"
            message = "Can't play checkers yet"
        }        
    }
}