
plan:
	rm -f .terraform.lock.hcl && terraform init && TF_LOG=INFO terraform plan -var-file=vars.tfvars

apply: plan
	TF_LOG=INFO terraform apply -var-file=vars.tfvars --auto-approve

destroy: plan
	TF_LOG=INFO terraform destroy -var-file=vars.tfvars --auto-approve