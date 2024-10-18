 variable "database" {
  type = connection.steampipe
  description = "The database connection to use."
  default = connection.steampipe.mock
}

variable "notifier" {
  type = notifier
  description = "The notifier to use."
  default = notifier.default
}