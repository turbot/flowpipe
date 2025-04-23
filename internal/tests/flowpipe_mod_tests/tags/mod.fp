mod "tags" {

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
}