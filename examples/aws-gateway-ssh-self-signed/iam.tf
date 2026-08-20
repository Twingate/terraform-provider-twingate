# Instance profile created by SSM Quick Setup — gives EC2 instances Session
# Manager access without a public IP or open SSH port.
#
# Prerequisites: run SSM Quick Setup in your AWS account first.
# https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-quick-setup.html

locals {
  instance_profile_name = "AmazonSSMRoleForInstancesQuickSetup"
}
