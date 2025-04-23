variable "slack_token" {
    type = string
    default = "foo"
}

variable "with_enum" {
    type = string
    enum = ["foo", "bar"]
    default = "foo"
}

// variable "connection_type" {
//     type = connection.aws
//     default = connection.aws.default
// }
