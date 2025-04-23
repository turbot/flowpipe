mod "mod_with_dynamic_conn" {
  title = "mod_with_dynamic_conn"
}


pipeline "conn_aws" {
    param "conn" {
        type    = string
        default = "aws_static"
    }

    param "cred_2" {
        type    = string
        default = "aws_static"
    }

    step "transform" "aws" {
        value   = connection.aws[param.conn].env
    }

    step "transform" "aws_access_key" {
        value = connection.aws[param.conn].access_key
    }

    step "transform" "aws_access_key_combo" {
        value = "access key: ${connection.aws[param.conn].access_key} and secret key is: ${connection.aws[param.cred_2].secret_key}"
    }    

    output "val" {
        value = step.transform.aws.value
    }

    output "val_access_key" {
        value = step.transform.aws_access_key.value
    }
}


pipeline "dynamic_conn_in_output" {

    param "conn" {
        type = string
        default = "example"
    }

    step "transform" "test" {
        output "val" {
            value = connection.aws[param.conn]
        }
    }

    output "val" {
        value = step.transform.test.output.val.access_key
    }
}
