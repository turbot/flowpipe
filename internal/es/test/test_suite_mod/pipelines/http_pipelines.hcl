pipeline "http_post_url_encoded" {

  step "http" "post" {
    title  = "Google OAuth Authorization"
    url    = "https://google.com"
    method = "POST"

    request_headers = {
      Content-Type = "application/x-www-form-urlencoded"
    }

    request_body = "code=4/0AfJohXnevkkXCjo_M8YPvLvoSxFVsNneFDkkQm9vFYJ-setwwvdqtupRuN-nrIjioG2H6Q&client_id=979620418102-k3c0nkk1g0t3k569m9nf15rmtsuofkg0.apps.googleusercontent.com&client_secret=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&redirect_uri=http://localhost&grant_type=authorization_code"
  }
}