pipeline "cred_aws" {
    param "cred" {
        type    = string
        default = "aws_static"
    }

    step "transform" "aws" {
        value   = credential.aws[param.cred].env
    }

    step "transform" "aws_access_key" {
        value = credential.aws[param.cred].access_key
    }

    output "val" {
        value = step.transform.aws.value
    }

    output "val_access_key" {
        value = step.transform.aws_access_key.value
    }
}

pipeline "cred_gcp" {
    param "cred" {
        type    = string
        default = "default"
    }

    step "transform" "gcp" {
        value   = credential.gcp[param.cred].access_token
    }

    output "val" {
        value = step.transform.gcp.value
    }
}


pipeline "cred_basic" {
    param "cred" {
        type    = string
        default = "credentials"
    }

    step "transform" "basic_username" {
        value   = credential.basic[param.cred].username
    }

    step "transform" "basic_password" {
        value   = credential.basic[param.cred].password
    }

    output "val_username" {
        value = step.transform.basic_username.value
    }

    output "val_password" {
        value = step.transform.basic_password.value
    }
}

pipeline "cred_slack" {
    param "cred" {
        type    = string
        default = "default"
    }

    param "null_param" {
        type = string
        optional = true
    }

    step "transform" "token" {
        value   = credential.slack[param.cred].token
    }

    output "slack_token" {
        value = step.transform.token.value
    }
}

pipeline "cred_gitlab" {
    param "cred" {
        type    = string
        default = "default"
    }

    param "null_param" {
        type = string
        optional = true
    }

    step "transform" "token" {
        value   = credential.gitlab[param.cred].token
    }

    output "gitlab_token" {
        value = step.transform.token.value
    }
}

pipeline "cred_abuseipdb" {
    param "cred" {
        type    = string
        default = "default"
    }

    param "null_param" {
        type = string
        optional = true
    }

    step "transform" "api_key" {
        value   = credential.abuseipdb[param.cred].api_key
    }

    output "abuseipdb_api_key" {
        value = step.transform.api_key.value
    }
}

pipeline "cred_clickup" {
    param "cred" {
        type    = string
        default = "default"
    }

    param "null_param" {
        type = string
        optional = true
    }

    step "transform" "token" {
        value   = credential.clickup[param.cred].token
    }

    output "clickup_token" {
        value = step.transform.token.value
    }
}

pipeline "multiple_credentials" {

    param "default_cred" {
        type    = string
        default = "default"
    }

    param "slack_cred" {
        type    = string
        default = "slack_static"
    }

     param "gitlab_cred" {
        type    = string
        default = "gitlab_static"
    }

    // slack
    step "transform" "slack_token" {
        value = credential.slack[param.slack_cred].token
    }

    output "slack_token_val" {
        value = step.transform.slack_token.value
    }

    // slack default
    step "transform" "default_slack_token" {
        value   = credential.slack[param.default_cred].token
    }

    output "slack_default_token" {
        value = step.transform.default_slack_token.value
    }

    // gitlab
     step "transform" "gitlab_token" {
        value = credential.gitlab[param.gitlab_cred].token
    }

    output "gitlab_token_val" {
        value = step.transform.gitlab_token.value
    }

    // gitlab default
    step "transform" "default_gitlab_token" {
        value   = credential.gitlab[param.default_cred].token
    }

    output "gitlab_default_token" {
        value = step.transform.default_gitlab_token.value
    }

    // clickup
    step "transform" "clickup_token" {
        value   = credential.clickup[param.default_cred].token
    }

    output "clickup_token_val" {
        value = step.transform.clickup_token.value
    }
}
