pipeline "param_test" {
    param "simple" {
        type = string
        default = "foo"
    }

    param "map_user_data" {
        type = map(string)
        default = {
            first_name = "felix"
            last_name = "mendelssohn"
            nationality = "german"
        }
    }

    param "object_diff_types" {
        type = any
        default = {
            string = "string"
            number = 1
            bool = true
            list = ["a", "b", "c"]
            map = {
                a = "a"
                b = "b"
                c = "c"
            }
        }
    }

    param "list_band_names" {
        type = list(string)
        default = [
            "Green Day",
            "New Found Glory",
            "Sum 41",
            "Blink 182",
            "Bowling for Soup",
            "My Chemical Romance",
            "The All-American Rejects",
        ]
    }

    step "transform" "simple" {
        value = param.simple
    }

    step "transform" "map_echo" {
        value = param.map_user_data.first_name
    }

    step "transform" "for_with_list" {
        for_each = param.list_band_names
        value    = each.value
    }

    step "transform" "for_with_list_and_index" {
        for_each = param.list_band_names
        value    = "${each.key}: ${each.value}"
    }

    step transform "map_diff_types_string" {
        value = param.object_diff_types.string
    }

    step transform "map_diff_types_number" {
        value = param.object_diff_types.number
    }

    step "transform" "for_each_list_within_map" {
        for_each = param.object_diff_types.list
        value    = each.value
    }
}

pipeline "param_override_test" {
    param "simple" {
        type = string
        default = "foo"
    }


    step "transform" "simple" {
        value = param.simple
    }
}