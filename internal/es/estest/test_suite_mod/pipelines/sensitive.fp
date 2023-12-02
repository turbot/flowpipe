pipeline "sensitive_one" {

    step "transform" "one" {

        value = {
            "one" = "two"
            "AWS_ACCESS_KEY_ID" = "abc"
            "AWS_SECRET_ACCESS_KEY" = "def"

            "pattern_match_aws_access_key_id" = "AKIAFAKEFAKEFAKEFAKE"
            "close_but_no_cigar" = "AKFFFAKEFAKEFAKEFAKE"

            "facebook_access_token" = "EAACEdEose0cBA1234FAKE1234"
        }
    }

    output "val" {
        value = step.transform.one.value
    }
}