
init:
#	@if [[ -f .terraform.lock.hcl ]]; then \
#		rm -f .terraform.lock.hcl && terraform init; \
#  	fi

plan: init
	@terraform plan -var-file=vars.tfvars

apply: init
	@terraform apply -var-file=vars.tfvars --auto-approve

destroy: init
	@terraform destroy -var-file=vars.tfvars --auto-approve