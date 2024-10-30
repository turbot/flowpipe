mod "test" {

}

pipeline "lots_of_sleep_on_pipeline" {
    max_concurrency = 2

    step "sleep" "lots_of_them" {
        for_each = [1, 2, 3]
        max_concurrency = 3
        
        duration = "5s"
    }
}
