connection "vault_1" {
  plugin    = "theapsgroup/vault"
  address   = "https://vault.mycorp.com/"
  auth_type = "token"
  token     = "sometoken"
}

connection "vault_2" {
  plugin       = "theapsgroup/vault"
  address      = "https://vault.mycorp.com/"
  auth_type    = "aws"
  aws_role     = "steampipe-test-role"
  aws_provider = "aws"
}