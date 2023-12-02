
// credential "aws" "aws_static" {
//     access_key = "ASIAQGDFAKEKGUI5MCEU"
//     secret_key = "QhLNLGM5MBkXiZm2k2tfake+TduEaCkCdpCSLl6U"
// }

// credential "basic" "credentials" {
//     username = "foo"
//     password = "bar"
// }


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

