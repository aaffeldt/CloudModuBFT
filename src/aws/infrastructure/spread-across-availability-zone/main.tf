
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-strategies.html#placement-groups-spread

# Spread placement groups

# A spread placement group is a group of instances that are each placed on distinct hardware.

# Spread placement groups are recommended for applications that have a small number of critical instances that should be kept separate from each other. Launching instances in a spread level placement group reduces the risk of simultaneous failures that might occur when instances share the same equipment. Spread level placement groups provide access to distinct hardware, and are therefore suitable for mixing instance types or launching instances over time.

# If you start or launch an instance in a spread placement group and there is insufficient unique hardware to fulfill the request, the request fails. Amazon EC2 makes more distinct hardware available over time, so you can try your request again later. Placement groups can spread instances across racks or hosts. Rack level spread placement groups can be used in AWS Regions and on AWS Outposts. Host level spread placement groups can be used with AWS Outposts only.

# Rack level spread placement groups
# The following image shows seven instances in a single Availability Zone that are placed into a spread placement group. The seven instances are placed on seven different racks, each rack has its own network and power source.

# A spread placement group.

# A rack level spread placement group can span multiple Availability Zones in the same Region. In a Region, a rack level spread placement group can have a maximum of seven running instances per Availability Zone per group. With Outposts, a rack level spread placement group can hold as many instances as you have racks in your Outpost deployment.

# Host level spread placement groups
# Host level spread placement groups are only available with AWS Outposts. A host spread level placement group can hold as many instances as you have hosts in your Outpost deployment. For more information, see Placement groups on AWS Outposts.

# Rules and limitations
# The following rules apply to spread placement groups:

# A rack spread placement group supports a maximum of seven running instances per Availability Zone. For example, in a Region with three Availability Zones, you can run a total of 21 instances in the group, with seven instances in each Availability Zone. If you try to start an eighth instance in the same Availability Zone and in the same spread placement group, the instance will not launch. If you need more than seven instances in an Availability Zone, we recommend that you use multiple spread placement groups. Using multiple spread placement groups does not provide guarantees about the spread of instances between groups, but it does help ensure the spread for each group, thus limiting the impact from certain classes of failures.

# Spread placement groups are not supported for Dedicated Instances.

# Host level spread placement groups are only supported for placement groups on AWS Outposts. A host level spread placement group can hold as many instances as you have hosts in your Outpost deployment.

# In a Region, a rack level spread placement group can have a maximum of seven running instances per Availability Zone per group. With AWS Outposts, a rack level spread placement group can hold as many instances as you have racks in your Outpost deployment.

# Capacity Reservations do not reserve capacity in a spread placement group.


terraform {

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "The AWS region to create things in"
  type        = string
  default     = "eu-central-1"

}
variable "ami" {
  description = "The AMI to use for the instance"
  type        = string
  default     = "ami-0a116fa7c861dd5f9"
}
variable "instance_type" {
  description = "The type of instance to start"
  type        = string
  default     = "t3.nano"

}
variable "client_instance_type" {
  description = "The type of instance to start for clients"
  type        = string
}

variable "number_of_nodes" {
  description = "Number of nodes to create"
  type        = number
}
variable "number_of_clients" {
  description = "Number of clients to create"
  type        = number
}

variable "number_of_node_spread_groups" {
  description = "Number of clusters to create"
  type        = number
}

variable "list_of_node_azs" {
  description = "List of availability zones to use"
  type        = string
  #   default     = ["a","b","c","a","b","c"]
}
variable "list_of_client_azs" {
  description = "List of availability zones to use"
  type        = string
  #   default     = ["a","b","c","a","b","c"]
}

variable "list_of_node_spread_group" {
  description = "List of spread groups to use"
  type        = list(number)
  #   default     = [0, 1, 2, 3, 4, 5]

}
# variable "list_of_client_spread_group" {
#   description = "List of spread groups to use"
#   type        = list(number)
#   #   default     = [0, 1, 2, 3, 4, 5]

# }

variable "client_az_ids" {
  description = "List of availability zone ids to use for clients"
  type        = list(number)
  #   default     = ["eu-central-1a","eu-central-1b","eu-central-1c","eu-central-1a","eu-central-1b","eu-central-1c"]
}

variable "node_az_ids" {
  description = "List of availability zone ids to use for nodes"
  type        = list(number)
  #   default     = ["eu-central-1a","eu-central-1b","eu-central-1c","eu-central-1a","eu-central-1b","eu-central-1c"]

}
variable "public_key_path" {
  description = "Path to the public key for controller instance"
  type        = string

}


# locals {
#   node_ip_addresses   = slice(module.vpc.ip_addresses, 0, var.number_of_nodes)
#   client_ip_addresses = slice(module.vpc.ip_addresses, var.number_of_nodes, var.number_of_nodes + var.number_of_clients)
# }

resource "aws_placement_group" "spread_across_az" {
  count    = var.number_of_node_spread_groups
  name     = "bft-placement-group-${count.index}"
  strategy = "spread"
}

# split var.list_of_node_azs
locals {
  list_of_node_azs = split(",", var.list_of_node_azs)
  list_of_client_azs = split(",", var.list_of_client_azs)
}


module "vpc" {
  source      = "../modules/spread-vpc"
  region      = var.region
  cidr_block  = "12.0.0.0/22"
  list_of_azs = sort(distinct(concat(local.list_of_node_azs , local.list_of_client_azs)))
}

output "azs_ip_adresses" {
  value = module.vpc.azs_ip_adresses
}


module "controller" {
  source                  = "../modules/spread-controller"
  vpc_cidr_block          = module.vpc.vpc_cidr_block
  control_instance_eip_id = module.vpc.control_instance_eip_id
  public_key_path         = var.public_key_path
  ami                     = var.ami
  security_group_ids      = [module.vpc.security_group_ids["allow_ssh_from_anywhere"], module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  subnet                  = module.vpc.controller_subnet.id
  availability_zone       = module.vpc.controller_az
  private_ip              = module.vpc.controller_ip
}

output "placement_group_names" {
  value = aws_placement_group.spread_across_az[*].name
}
output "list_of_node_spread_group" {
  # list_of_node_spread_group
  value = var.list_of_node_spread_group
}
output "placement_group_ids" {
  #forloop
  # value = aws_placement_group.spread_across_az[var.list_of_node_spread_group[count.index]].id
  value = {
    for idx, pg in var.list_of_node_spread_group :
    idx => aws_placement_group.spread_across_az[pg].id
  }
}
output "security_group_id_any_data_traffic_in_subnet" {
  # [module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  value = module.vpc.security_group_ids["any_data_traffic_in_subnet"]
}

locals {
  node_ip_addresses   = [for idx, az in local.list_of_node_azs : module.vpc.azs_ip_adresses[az][var.node_az_ids[idx]]]
  client_ip_addresses = [for idx, az in local.list_of_client_azs : module.vpc.azs_ip_adresses[az][var.client_az_ids[idx]]]
}

output "ip_addresses" {
  value = {
    nodes   = join(",", local.node_ip_addresses)
    clients = join(",", local.client_ip_addresses)
  }
}



module "nodes" {
  count              = var.number_of_nodes
  source             = "../modules/spread-node"
  ami                = var.ami
  instance_type      = var.instance_type
  private_ip         = local.node_ip_addresses[count.index]
  security_group_ids = [module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  ssh_key_name       = module.controller.key_pair_name
  placement_group_id = aws_placement_group.spread_across_az[var.list_of_node_spread_group[count.index]].id
  availability_zone  = local.list_of_node_azs[count.index]
  subnet_id          = module.vpc.subnets[local.list_of_node_azs[count.index]].id
}




module "clients" {
  count              = var.number_of_clients
  source             = "../modules/spread-node"
  ami                = var.ami
  instance_type      = var.client_instance_type
  private_ip         = local.client_ip_addresses[count.index]
  security_group_ids = [module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  ssh_key_name       = module.controller.key_pair_name
  # placement_group_id = aws_placement_group.spread_across_az[var.list_of_client_spread_group[count.index]].id
  availability_zone  = local.list_of_client_azs[count.index]
  subnet_id          = module.vpc.subnets[local.list_of_client_azs[count.index]].id
}



# output "number_of_possible_ips" {
#   value = module.vpc.cidr_block_number
# }

# output "ip_addresses" {
#   value = {
#     nodes   = join(",", local.node_ip_addresses)
#     clients = join(",", local.client_ip_addresses)
#   }
# }

output "node_ip_addresses" {
  value = join(",", local.node_ip_addresses)
}
output "client_ip_addresses" {
  value = join(",", local.client_ip_addresses)
}

output "controller_ip" {
  value = module.vpc.control_instance_eip_ip
}
