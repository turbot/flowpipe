pipeline "reduce_list" {
    param "input" {
        type = list(number)
        default = [1, 2, 3, 4, 5, 6]
    }

    step "transform" "echo" {
        for_each = param.input
        if       = each.value % 2 == 0
        value    = "${each.key}: ${each.value}"
    }

    output "val_1" {
        value = step.transform.echo[1].value
    }

    output "val" {
       value = step.transform.echo
    }
 }

 pipeline "reduce_map" {
    param "input" {
        type = map(any)
        default = {
            "green_day" = {
                "name" = "Green Day"
                "albums" = ["Dookie", "American Idiot", "Nimrod"]
            },
            "blink_182" = {
                "name" = "Blink 182"
                "albums" = ["Enema of the State", "Take Off Your Pants and Jacket", "California"]
            },
            "sum_41" = {
                "name" = "Sum 41"
                "albums" = ["All Killer No Filler", "Does This Look Infected?", "Chuck"]
            }
        }
    }

    step "transform" "echo" {
        for_each  = param.input
        if        = each.key != "blink_182"
        value     = "${each.key}: ${each.value.name}"
    }

    output "val" {
        value = step.transform.echo
    }

     output "val_two" {
        value = step.transform.echo["sum_41"].value
    }
 }