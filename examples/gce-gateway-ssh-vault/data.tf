data "terraform_remote_state" "vault" {
  backend = "local"

  config = {
    path = "${path.module}/vault/terraform.tfstate"
  }
}
