pipeline "lots_of_sleep" {

    step "sleep" "lots_of_them" {

        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
        duration = "5s"
    }
}


pipeline "lots_of_sleep_bound" {
    step "sleep" "lots_of_them" {
        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9]
        max_concurrency = 3
        duration = "5s"
    }
}


pipeline "lots_of_sleep_on_pipeline" {
    max_concurrency = 2

    step "sleep" "lots_of_them" {
        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9]
        max_concurrency = 3
        duration = "5s"
    }

}
