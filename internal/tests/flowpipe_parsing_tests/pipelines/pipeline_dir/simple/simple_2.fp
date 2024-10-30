pipeline "simple_http_file_2" {
    description = "my simple http pipeline in second file"
    step "http" "my_step_1" {
        url = "http://localhost:8081"
    }
}