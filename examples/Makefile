
init:
	@if [[ -f .terraform.lock.hcl ]]; then \
		rm -f .terraform.lock.hcl && terraform init; \
  	fi

plan:
	@TF_LOG=info terraform plan -var-file=vars.tfvars

apply:
	@TF_LOG=info terraform apply -var-file=vars.tfvars --auto-approve

destroy:
	@TF_LOG=info terraform destroy -var-file=vars.tfvars --auto-approve