pipeline "reduce_list" {
    param "input" {
        type = list(number)
        default = [1, 2, 3, 4, 5, 6]
    }

    step "echo" "echo" {
        for_each = param.input
        if = each.value % 2 == 0
        text = "${each.key}: ${each.value}"
    }

    output "val_1" {
        value = step.echo.echo[1].text
    }

    output "val" {
       value = step.echo.echo
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

    step "echo" "echo" {
        for_each = param.input
        if = each.key != "blink_182"
        text = "${each.key}: ${each.value.name}"
    }

    output "val" {
        value = step.echo.echo
    }

     output "val_two" {
        value = step.echo.echo["sum_41"].text
    }
 }