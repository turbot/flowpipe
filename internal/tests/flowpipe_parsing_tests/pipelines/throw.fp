pipeline "throw_simple_no_unresolved" {
    step "transform" "one" {
        value = "foo"

        throw {
            if = true
            message = "foo"
        }
    }
}

pipeline "throw_simple_unresolved" {
    step "transform" "one" {
        value = "foo"

        throw {
            if = result.value == "foo"
            message = "foo"
        }
    }
}


pipeline "throw_multiple" {
    step "transform" "base" {
        value = "bar"
    }

    step "transform" "base_2" {
        value = "bar"
    }    

    step "transform" "one" {
        value = "foo"

        throw {
            if = result.value == "foo"
            message = step.transform.base.value
        }

        throw {
            if = true
            message = step.transform.base_2.value
        }

        throw {
            if = result.value == "foo"
            message = "baz"
        }

        throw {
            if = false
            message = "qux"
        }

    }
}

