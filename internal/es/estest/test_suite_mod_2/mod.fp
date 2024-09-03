mod "test_suite_mod_2" {
    title = "Test Suite Mod 2"
}

pipeline "with_tags" {
    title = "tags on pipeline"

    tags = {
        "tag1" = "value1"
        "tag2" = "value2"
    }

    param "tag_param" {
        type = string
        default = "value"
        tags = {
            "tag3" = "value3"
            "tag4" = "value4"
        }
    }

    param "list_param" {
        type = list(string)
        default = ["value1", "value2"]
        tags = {
            "tag5" = "value5"
            "tag6" = "value6"
        }
    }
}