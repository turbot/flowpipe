pipeline "defaul_aws_credential" {

    param "region" {
        default = "us-east-1"
    }

    param "cred" {
        default = "default"
    }

    step "transform" "transform" {
        value = merge(credential.aws[param.cred].env, { AWS_REGION = param.region })
    }

    output "val" {
        value = step.transform.transform
    }
}