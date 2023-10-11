function "hello_python" {
    runtime = "python:3.10"
    handler = "app.my_handler"
    src = "./functions/hello-python"
    env = {
        "HELLO": "world"
    }
}