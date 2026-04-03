output "instance_id" {
  value = aws_instance.bft_instance.id
}

output "private_ip" {
  value = aws_instance.bft_instance.private_ip
}

output "public_ip" {
  value = aws_instance.bft_instance.public_ip
}

variable "ami" {
  description = "AMI ID for the instances"
  type        = string
}

variable "instance_type" {
  description = "Instance type for the EC2 instances"
  type        = string
}

variable "private_ip" {
  description = "Private IP address for the EC2 instance"
  type        = string
}

variable "security_group_ids" {
  description = "List of security group IDs to associate with the EC2 instance"
  type        = list(string)
}

variable "subnet_id" {
  description = "Subnet ID for the EC2 instance"
  type        = string
}

variable "ssh_key_name" {
  description = "Key pair name for the EC2 instance"
  type        = string
}

variable "placement_group_id" {
  description = "Placement group ID for the EC2 instance"
  type        = string
  default     = null
}

# variable "placement_partition_number" {
#   description = "Partition number for the placement group"
#   type        = number
#   default     = null # Set to null if not using a placement group
# }
variable "availability_zone" {
  description = "The availability zone for the instance"
  type        = string
}

resource "aws_instance" "bft_instance" {
  #   count         = var.number_of_instances
  ami               = var.ami
  instance_type     = var.instance_type
  key_name          = var.ssh_key_name
  private_ip        = var.private_ip
  security_groups   = var.security_group_ids # Use the provided security group IDs
  placement_group   = var.placement_group_id
  availability_zone = var.availability_zone
  subnet_id         = var.subnet_id

  tags = {
    Name = "bft-instance-${var.private_ip}"
  }
  provisioner "local-exec" { command = "aws ec2 wait instance-status-ok --instance-ids ${self.id} --region=eu-central-1" }
}


