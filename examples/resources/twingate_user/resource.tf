provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_user" "user" {
  email = "sample@company.com"
  first_name = "Twin"
  last_name = "Gate"
  role = "DEVOPS"
  send_invite = true
}
