output "number_of_possible_ips" {
  value = module.vpc.cidr_block_number
}

output "ip_addresses" {
  value = {
    nodes = join(",", local.node_ip_addresses)
    clients   = join(",", local.client_ip_addresses)
  }
}

output "node_ip_addresses" {
  value = join(",", local.node_ip_addresses)
}
output "client_ip_addresses" {
  value = join(",", local.client_ip_addresses)
}

output "controller_ip" {
  value = module.vpc.control_instance_eip_ip
}

variable "region" {
  description = "The AWS region to deploy resources in"
  type        = string
  default     = "eu-central-1"
}

variable "availability_zone" {
  description = "The availability zone to deploy resources in"
  type        = string
  default     = "eu-central-1a"
}

variable "public_key_path" {
  description = "Path to the public key for controller instance"
  type        = string
}

variable "number_of_partitions" {
  description = "Number of partitions to use in the placement group"
  type = number
}

variable "number_of_nodes" {
  description = "Total number of nodes (leaders + followers)"
  type        = number  
}

variable "number_of_clients" {
  description = "Total number of client nodes"
  type        = number  
}

variable "list_of_node_partitions" {
  description = "List of partitions to use for the placement group"
  type        = list(number)  
}
variable "list_of_client_partitions" {
  description = "List of partitions to use for the placement group"
  type        = list(number)  
}

variable "ami" {
  description = "AMI ID for the instances"
  type        = string
}

variable "instance_type" {
  description = "Instance type for the EC2 instances"
  type        = string
}

variable "client_instance_type" {
  description = "Instance type for the client EC2 instances"
  type        = string
}

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

resource "aws_placement_group" "placement_group" {
  name="bft-placement-group"
  strategy = "partition"
  partition_count = var.number_of_partitions
}


module "vpc" {
  source            = "../modules/vpc"
  region            = var.region
  availability_zone = var.availability_zone
  cidr_block        = "10.0.0.0/24"
}


module "ssh_key" {
  source = "../modules/key"
}


module "controller" {
  source                  = "../modules/controller"
  vpc_cidr_block          = module.vpc.vpc_cidr_block
  subnet_id               = module.vpc.subnet_id
  control_instance_eip_id = module.vpc.control_instance_eip_id
  public_key_path         = var.public_key_path
  ami                     = var.ami
  security_group_ids      = [module.vpc.security_group_ids["allow_ssh_from_anywhere"], module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
}

locals {
  node_ip_addresses = slice(module.vpc.ip_addresses,0,var.number_of_nodes)
  client_ip_addresses = slice(module.vpc.ip_addresses,var.number_of_nodes,var.number_of_nodes+var.number_of_clients)
}

module "nodes" {
  count =  var.number_of_nodes
  source = "../modules/node"
  ami = var.ami 
  instance_type =  var.instance_type
  private_ip =  local.node_ip_addresses[count.index]
  security_group_ids                 = [module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  subnet_id                          = module.vpc.subnet_id
  ssh_key_name   = module.controller.key_pair_name
  placement_group_id                 = aws_placement_group.placement_group.id
  placement_partition_number         = var.list_of_node_partitions[count.index]  
}


module "clients" {
  count                              = var.number_of_clients
  source                             = "../modules/node"
  ami                                = var.ami
  instance_type                      = var.client_instance_type
  private_ip                         = local.client_ip_addresses[count.index]
  security_group_ids                 = [module.vpc.security_group_ids["any_data_traffic_in_subnet"]]
  subnet_id                          = module.vpc.subnet_id
  ssh_key_name  = module.controller.key_pair_name
  placement_group_id                 = aws_placement_group.placement_group.id
  placement_partition_number         = var.list_of_client_partitions[count.index] 
}



