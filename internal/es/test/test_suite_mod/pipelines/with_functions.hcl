pipeline "with_functions" {

    step "function" "hello_nodejs_step" {
        function = function.hello_nodejs
    }

    output "val" {
        value = step.function.hello_nodejs_step.result
    }

}