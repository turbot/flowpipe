pipeline "lots_of_sleep" {

    step "sleep" "lots_of_them" {

        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
        duration = "5s"
    }
}
