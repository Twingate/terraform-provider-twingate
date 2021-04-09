
plan:
	rm .terraform.lock.hcl && terraform init && TF_LOG=INFO terraform plan -var-file=vars.tfvars
apply:
	TF_LOG=INFO terraform apply -var-file=vars.tfvars --auto-approve
