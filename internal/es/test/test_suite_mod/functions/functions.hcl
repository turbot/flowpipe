function "hello_python" {
    runtime = "python:3.10"
    handler = "app.my_handler"
    src = "./hello-python/"
    env = {
        "HELLO": "world"
    }
}