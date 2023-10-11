pipeline "with_functions" {

    step "function" "hello_python_step" {
        function = function.hello_python
    }

}