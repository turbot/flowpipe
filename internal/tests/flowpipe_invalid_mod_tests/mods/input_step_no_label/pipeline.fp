pipeline "email_input" {
    step "input" "email" {
        notifier = notifier.default

        type   = "button"
        prompt = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

        channel = "#flowpipe-core"

        to = ["victor+one@turbot.com", "jonah.hadianto@gmail.com"]

        option {}
        option "Deny" {}
    }
}
