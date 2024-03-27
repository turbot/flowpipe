pipeline "lots_of_sleep" {

    step "sleep" "lots_of_them" {

        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
        duration = "1s"
    }
}


pipeline "lots_of_sleep_bound" {
    step "sleep" "lots_of_them" {
        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9]
        max_concurrency = 3
        duration = "5s"
    }
}

pipeline "lots_of_nested_pipeline" {
    step "pipeline" "pipeline" {
        for_each = [1, 2, 3, 4, 5, 6, 7]
        max_concurrency = 2
        pipeline = pipeline.nested
    }
}

pipeline "nested" {
    step "transform" "transform" {
        value = "foo"
    }
}


pipeline "lots_of_sleep_bound_with_param" {
    param "concurrency" {
        default = 1
    }
    step "sleep" "lots_of_them" {
        for_each = [1, 2, 3, 4, 5, 6, 7, 8, 9]
        max_concurrency = param.concurrency + 1
        duration = "2s"
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
