
init:
	@if [[ -f .terraform.lock.hcl ]]; then \
		rm -f .terraform.lock.hcl && terraform init; \
  	fi

plan: init
	TF_LOG=INFO terraform plan -var-file=vars.tfvars

apply: init
	TF_LOG=INFO terraform apply -var-file=vars.tfvars --auto-approve

destroy: init
	TF_LOG=INFO terraform destroy -var-file=vars.tfvars --auto-approve