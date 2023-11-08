mod "github_integrations" {
  title = "GitHub Integrations"
  description   = "Run GitHub mod pipelines along with other mods."
  color         = "#191717"

  require {
    mod "github.com/turbot/flowpipe-mod-github" {
      version = "*"
      args = {
        repository_full_name = "turbot/steampipe"
        access_token                = "TOKEN"
      }
    }

    mod "github.com/turbot/flowpipe-mod-slack" {
      version = "*"
      args = {
        token   = "turbot/steampipe"
        channel   = "TOKEN"
      }
    }
  }
}
#
## TODO: Remove defaults once the bug in dependency mods is fixed
#variable "github_token" {
#  type        = string
#  description = "The GitHub personal access token to authenticate to the GitHub APIs."
#  default     = "FOFOFOFOFOF"
#}
#
#variable "github_repository_full_name" {
#  type        = string
#  description = "The full name of the GitHub repository. Examples: turbot/steampipe, turbot/flowpipe"
#  default = "turbot/steampipe"
#}
