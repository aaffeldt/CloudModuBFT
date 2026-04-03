# VPC module (vpc.tf)

Purpose
- Creates a single VPC, subnet, route table, internet gateway, security groups, an Elastic IP, and an S3 VPC endpoint.
- Intended for small test or demo deployments of the CloudModuBFT network.

Provided Terraform resources (key items)
- aws_vpc.vpc
- aws_subnet.subnet
- aws_route_table.bft_route_table
- aws_route_table_association.btf_route_table_association
- aws_internet_gateway.gateway
- aws_route.internet_access
- aws_eip.control_instance_eip
- aws_security_group.allow_ssh_from_anywhere
- aws_security_group.any_data-traffic-in-subnet
- aws_vpc_endpoint.s3_endpoint

Inputs (variables)
- region (string): AWS region for endpoints and service names.
- availability_zone (string): AZ to place the subnet.
- cidr_block (string, default "10.0.0.0/24"): VPC CIDR block.
- allow_any_data_traffic_from_other_vpcs (list, default []): Extra CIDR blocks to allow in the "any data traffic" SG.

Outputs
- vpc_id
- vpc_cidr_block
- subnet_id
- control_instance_eip_id
- control_instance_eip_ip (Elastic IP public IP)
- cidr_block_number (local computed value)
- ip_addresses (local computed list)
- security_group_ids (map with keys: allow_ssh_from_anywhere, any_data_traffic_in_subnet)

Usage example (module call)
```hcl
module "vpc" {
  source                       = "./src/aws/infrastructure/modules/vpc"
  region                       = "us-east-1"
  availability_zone            = "us-east-1a"
  cidr_block                   = "10.0.0.0/24"
  allow_any_data_traffic_from_other_vpcs = ["10.1.0.0/16"]
}

output "control_ip" {
  value = module.vpc.control_instance_eip_ip
}
```

Notes and gotchas
- The module exposes an Elastic IP to be attached to control instances — capture `control_instance_eip_ip`.
- The security group `allow_ssh_from_anywhere` opens port 22 to 0.0.0.0/0; restrict in production.
- Terraform resource identifiers must use valid Terraform names (letters, digits and underscores). The current resource name `any_data-traffic-in-subnet` contains hyphens which can cause validation errors; consider renaming to `any_data_traffic_in_subnet` if you encounter issues.
- The module currently creates a single subnet equal to the VPC CIDR; adjust if you need multiple subnets or more advanced networking.

For further customization
- Add NAT, private/public subnet split, additional route tables, or more granular security group rules as needed for production deployments.
