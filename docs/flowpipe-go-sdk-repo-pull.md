# Flowpipe Go SDK repo pull in GitHub Actions


1. This is only valid while turbot/flowpipe-go-sdk repo is still a private repo
1. We use GitHub "Deploy Keys" to allow repo pull
1. VH generated Repo Keys.
1. Add the Public Key in `turbot/flowpipe-go-sdk` Deploy Keys repo.
1. Add the Private Key in `turbot/flowpipe` Actions Secret.
1. Works.

NOTE: this is a much safer (and recommended) option than setting a repo token.