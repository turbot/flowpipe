function "hello_python" {
    runtime = "python:3.10"
    handler = "app.my_handler"
    src = "./functions/hello-python"
    env = {
        "HELLO": "world"
    }
}

function "hello_nodejs" {
    runtime = "nodejs:18"
    handler = "index.handler"
    src = "./functions/hello-nodejs"
    env = {
        "HELLO": "world"
    }
}