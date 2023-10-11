function "hello-js" {
    runtime = "nodejs:18"
    src = "./lambda-nodejs"
    handler = "app.my_handler"
    env = {
        AWS_ACCESS_KEY_ID = "akia"
    }
}