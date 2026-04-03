# output "cidr_block_number" {
#   value = local.cidr_block_number

# }


output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "vpc_cidr_block" {
  value = aws_vpc.vpc.cidr_block
}



output "control_instance_eip_id" {
  value = aws_eip.control_instance_eip.id
}

output "control_instance_eip_ip" {
  value = aws_eip.control_instance_eip.public_ip
}

# Outputs the IDs of security groups created in this module.
# - allow_ssh_from_anywhere: Security group allowing SSH access from any IP address.
# - any_data_traffic_in_subnet: Security group allowing any data traffic within the subnet.
output "security_group_ids" {
  value = {
    allow_ssh_from_anywhere    = aws_security_group.allow_ssh_from_anywhere.id
    any_data_traffic_in_subnet = aws_security_group.any_data-traffic-in-subnet.id
  }
}

# variable "availability_zone" {
#   description = "The availability zone for the subnet"
#   type        = string
# }

variable "list_of_azs" {
  description = "List of availability zones to create spread placement groups in"
  type        = list(string)
}

variable "region" {
  description = "The AWS region to deploy resources in"
  type        = string
}

variable "cidr_block" {
  default = "12.0.0.0/24"
}

variable "allow_any_data_traffic_from_other_vpcs" {
  default = []
}

locals {
  distinct_azs         = sort(distinct(var.list_of_azs))
  cidr_block           = [for idx, az in var.list_of_azs : cidrsubnet(aws_vpc.vpc.cidr_block, 3, index(var.list_of_azs, az))] # Subnet CIDR block
  az_cidr_block_number = { for idx, az in var.list_of_azs : az => pow(2, 32 - tonumber(regex("/([0-9]+)", local.cidr_block[idx])[0])) - 15 }

  azs_cidr_blocks = { for idx, az in var.list_of_azs : az => cidrsubnet(aws_vpc.vpc.cidr_block, 3, idx) }
  azs_ip_adresses = { for az, cidr in local.azs_cidr_blocks : az => [for i in range(local.az_cidr_block_number[az]) : cidrhost(cidr, i + 15)] }
}
output "azs_cidr_blocks" {
  value = local.azs_cidr_blocks
}
output "azs_ip_adresses" {
  value = local.azs_ip_adresses
}
output "controller_ip" {
  value = [for i in range(local.az_cidr_block_number[local.distinct_azs[0]]) : cidrhost(local.azs_cidr_blocks[local.distinct_azs[0]], 10)][0]
}
output "controller_subnet" {
  value = aws_subnet.subnet[local.distinct_azs[0]]
}
output "controller_az" {
  value = local.distinct_azs[0]
}

resource "aws_vpc" "vpc" {
  cidr_block           = var.cidr_block # The CIDR block for the VPC, provided via variable
  enable_dns_support   = true           # Enables DNS resolution within the VPC
  enable_dns_hostnames = true           # Enables DNS hostnames for instances launched into the VPC
  tags = {
    Name = "bft-vpc",
    BFT  = true, # Tag for easy identification in AWS Console
  }

}

# #---------------------------
# # Subnet: VPC Subnet
# #---------------------------
# resource "aws_subnet" "subnet" {
#   vpc_id     = aws_vpc.vpc.id         # Reference to the VPC created above
#   cidr_block = aws_vpc.vpc.cidr_block # Subnet CIDR block (same as VPC for single subnet)
#   tags = {
#     Name = "bft-subnet" # Tag for identification
#   }
#   availability_zone = var.availability_zone
# }

resource "aws_subnet" "subnet" {

  for_each          = toset(local.distinct_azs)
  vpc_id            = aws_vpc.vpc.id
  availability_zone = each.value
  cidr_block        = cidrsubnet(aws_vpc.vpc.cidr_block, 3, index(var.list_of_azs, each.key)) # Subnet CIDR block
  tags = {
    Name = "bft-subnet-${each.key}" # Tag for identification
  }
}

#output a map of az => subnetid 
output "subnets" {
  value = aws_subnet.subnet
}


resource "aws_route_table" "bft_route_table" {
  vpc_id = aws_vpc.vpc.id # Attach to the created VPC

  tags = {
    Name = "bft-route-table" # Tag for identification
  }

}

resource "aws_route_table_association" "btf_route_table_association" {
  for_each       = aws_subnet.subnet
  subnet_id      = each.value.id                      # Subnet to associate
  route_table_id = aws_route_table.bft_route_table.id # Route table to associate
}
# resource "aws_route_table_association" "btf_route_table_association" {
#   subnet_id      = aws_subnet.subnet.id               # Subnet to associate
#   route_table_id = aws_route_table.bft_route_table.id # Route table to associate
# }


resource "aws_internet_gateway" "gateway" {
  vpc_id = aws_vpc.vpc.id # Attach the internet gateway to the VPC

  tags = {
    Name = "bft-internet-gateway" # Tag for identification
  }

}

resource "aws_route" "internet_access" {
  route_table_id         = aws_route_table.bft_route_table.id # Route table to update
  destination_cidr_block = "0.0.0.0/0"                        # Destination for internet access
  gateway_id             = aws_internet_gateway.gateway.id    # Use the internet gateway for routing
}


resource "aws_eip" "control_instance_eip" {
  #   vpc = true # Allocate an Elastic IP for VPC use
  tags = {
    Name = "bft-control-instance-eip" # Tag for identification
  }

}



resource "aws_security_group" "allow_ssh_from_anywhere" {
  name        = "bft-allow-ssh-from-anywhere"    # Name of the security group
  description = "Allow SSH access from anywhere" # Description of the security group
  vpc_id      = aws_vpc.vpc.id                   # Attach to the created VPC

  ingress {
    from_port   = 0             # Allow SSH on port 22
    to_port     = 0             # Allow SSH on port 22
    protocol    = "tcp"         # Protocol for SSH
    cidr_blocks = ["0.0.0.0/0"] # Allow access from anywhere
  }
  egress {
    from_port   = 0    # Allow all outbound traffic
    to_port     = 0    # Allow all outbound traffic
    protocol    = "-1" # All protocols
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "any_data-traffic-in-subnet" {
  name        = "bft-any-data-traffic-in-subnet"   # Name of the security group
  description = "Allow all data traffic in subnet" # Description of the security group
  vpc_id      = aws_vpc.vpc.id                     # Attach to the created VPC

  ingress {
    from_port   = 0                                                                                         # Allow all inbound traffic
    to_port     = 0                                                                                         # Allow all inbound traffic
    protocol    = "-1"                                                                                      # All protocols
    cidr_blocks = concat([aws_vpc.vpc.cidr_block, "0.0.0.0/0"], var.allow_any_data_traffic_from_other_vpcs) # Allow traffic from the VPC CIDR block
  }
  egress {
    from_port   = 0                                                                                         # Allow all outbound traffic
    to_port     = 0                                                                                         # Allow all outbound traffic
    protocol    = "-1"                                                                                      # All protocols
    cidr_blocks = concat([aws_vpc.vpc.cidr_block, "0.0.0.0/0"], var.allow_any_data_traffic_from_other_vpcs) # Allow traffic to the VPC CIDR block
  }
}

# resource "aws_vpc_endpoint" "s3_endpoint" {
#   vpc_id          = aws_vpc.vpc.id                       # Attach the VPC endpoint to the created VPC
#   service_name    = "com.amazonaws.${var.region}.s3"     # AWS S3 service endpoint
#   route_table_ids = [aws_route_table.bft_route_table.id] # Associate with the route table

#   tags = {
#     Name = "bft-s3-endpoint" # Tag for identification
#   }

# }


