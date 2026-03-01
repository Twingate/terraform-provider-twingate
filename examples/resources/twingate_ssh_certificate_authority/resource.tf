provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_ssh_certificate_authority" "example" {
  name       = "My SSH CA"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBN9QDVJ4MYHLfObTSxO0aevAMHsaVqaWt7OHZQ3yDvR"
}

# example loading public key from a file
resource "twingate_ssh_certificate_authority" "example_from_file" {
  name       = "My SSH CA from file"
  public_key = trimspace(file("${path.module}/keys/ca.pub"))
}
