function "hello-js" {
    runtime = "nodejs:18"
    src = "./lambda-nodejs"
    env = {
        AWS_ACCESS_KEY_ID = "akia"
    }
}