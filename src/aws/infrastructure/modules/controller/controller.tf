


variable "public_key_path" {
  description = "Path to the public key file for the control instance"
  type        = string
}

variable "ami" {
  description = "The AMI ID for the control instance"
  type        = string
}



variable "security_group_ids" {
  description = "List of security group IDs to associate with the control instance"
  type        = list(string)

}

variable "control_instance_eip_id" {
  description = "The allocation ID of the Elastic IP to associate with the control instance"
  type        = string
}

variable "vpc_cidr_block" {
  description = "The CIDR block of the VPC where the control instance will be deployed"
  type        = string
}

variable "subnet_id" {
  description = "The subnet ID where the control instance will be deployed"
  type        = string
}


output "key_pair_name" {
  value = aws_key_pair.control_instance_key_pair.key_name
  
}


resource "aws_key_pair" "control_instance_key_pair" {
  key_name = "bft-control-instance-key" # Name for the key pair

  public_key = file("${var.public_key_path}")
  tags = {
    Name = "bft-control-instance-key-pair" # Tag for identification
  }

}
resource "aws_instance" "control_instance" {
  ami           = var.ami
  instance_type = "t3.nano"
  subnet_id     = var.subnet_id
  key_name      = aws_key_pair.control_instance_key_pair.key_name # Use the created key pair
  private_ip    = cidrhost(var.vpc_cidr_block, 10)                # Assign a private IP within the VPC CIDR block
  # security_groups = [ aws_security_group.allow_ssh_from_anywhere.id,
  # aws_security_group.any_data-traffic-in-subnet.id ]
  security_groups = var.security_group_ids # Use the provided security group IDs

}

resource "aws_eip_association" "control_instance_eip_association" {
  instance_id   = aws_instance.control_instance.id # Associate the Elastic IP with the control instance
  allocation_id = var.control_instance_eip_id      # Reference to the allocated Elastic IP
}



# output "control_instance_ip" {
#   value = aws_eip.control_instance_eip.public_ip
# }
output "control_instance" {
  value = {
    instance = aws_instance.control_instance
  }

}
