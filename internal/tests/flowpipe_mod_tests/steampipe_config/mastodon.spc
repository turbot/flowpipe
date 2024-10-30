connection "mastodon_1" {
  plugin = "mastodon"

  access_token = "FK1_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU"
  server       = "https://myserver.social"
}

connection "mastodon_2" {
  plugin = "mastodon"

  access_token = "FK2_gBrl7b9sPOSADhx61-fakezv9EDuMrXuc1AlcNU"
  server       = "https://myserver.social"
  app          = "elk.zone"
}