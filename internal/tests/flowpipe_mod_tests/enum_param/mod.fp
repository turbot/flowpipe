mod "enum_param" {
  title = "mod with enum param"
}

pipeline "enum_param" {
  param "cred" {
    type    = string
    default = "aws_static"
    enum  = ["aws_static", "aws_dynamic"]
  }
}