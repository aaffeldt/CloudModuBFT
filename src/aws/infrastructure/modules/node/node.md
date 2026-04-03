# Node module (node.tf)

Purpose
- Launches a single EC2 instance configured for use as a CloudModuBFT node.
- Simple, reusable module intended for test or demo deployments.

Resources created
- aws_instance.bft_instance

Module variables (inputs)
- ami (string): AMI ID to use for the instance.
- instance_type (string): EC2 instance type (e.g., "t3.medium").
- private_ip (string): Private IP address to assign to the instance (must be available in the target subnet).
- security_group_ids (list(string)): Security group IDs to associate with the instance.
- subnet_id (string): Subnet ID in which to launch the instance.
- ssh_key_name (string): Name of the EC2 key pair to associate for SSH access.
- placement_group_id (string, nullable): Optional placement group identifier (if you place instances into a group).
- placement_partition_number (number, nullable): Optional partition number for placement groups that support partitions.

Outputs
- This module does not define outputs by default. Add outputs in your root module if you need instance ID, public IP, etc.

Usage example
```hcl
module "node" {
  source                = "./src/aws/infrastructure/modules/node"
  ami                   = "ami-0123456789abcdef0"
  instance_type         = "t3.medium"
  private_ip            = "10.0.0.15"
  security_group_ids    = [module.vpc.security_group_ids.any_data_traffic_in_subnet] # or other SG IDs
  subnet_id             = module.vpc.subnet_id
  ssh_key_name          = "my-keypair"
  placement_group_id    = null
  placement_partition_number = null
}

output "node_instance_id" {
  value = module.node.aws_instance.bft_instance.id
}
```

Notes and gotchas
- Security groups vs VPC security group IDs:
  - In a VPC you should use vpc_security_group_ids on aws_instance (which accepts SG IDs).
  - The module's aws_instance currently sets the "security_groups" attribute; that attribute expects security group names for EC2-Classic. If you pass SG IDs in a VPC, switch the instance attribute to `vpc_security_group_ids = var.security_group_ids`.
- Private IP:
  - The provided private_ip must be within the subnet's CIDR and not already in use.
  - For dynamic allocation, omit private_ip and let AWS assign one.
- Placement groups:
  - If using placement groups, ensure the placement group exists and is in the same AZ as the subnet.
  - The placement_partition_number is only applicable for partition placement groups.
- Tagging:
  - Instances are tagged with Name = "bft-instance-${var.private_ip}" — consider adjusting for readability in environments where private_ip is null.
- Security:
  - Ensure security group rules are restrictive enough for production (SSH access, data plane rules).
- Extensibility:
  - Consider adding user_data, IAM role/instance profile, EBS volumes, or outputs (instance id, private/public IP) as needed for your deployment workflows.
