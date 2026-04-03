# Controller Module

This module provisions an AWS EC2 instance to act as the control node for CloudModuBFT. It handles key pair creation, instance deployment, and Elastic IP association.

## Features

- Creates an EC2 instance with a specified AMI and instance type.
- Generates an SSH key pair for secure access.
- Associates security groups for network control.
- Assigns a static Elastic IP to the instance.

## Inputs

| Name                   | Description                                               | Type         | Required |
|------------------------|-----------------------------------------------------------|--------------|----------|
| `public_key_path`      | Path to the public key file for SSH access                | `string`     | yes      |
| `ami`                  | AMI ID for the EC2 instance                              | `string`     | yes      |
| `security_group_ids`   | List of security group IDs to attach                      | `list(string)`| yes     |
| `control_instance_eip_id` | Allocation ID of the Elastic IP                       | `string`     | yes      |
| `vpc_cidr_block`       | CIDR block of the target VPC                             | `string`     | yes      |
| `subnet_id`            | Subnet ID for instance deployment                        | `string`     | yes      |

## Outputs

| Name                | Description                                 |
|---------------------|---------------------------------------------|
| `key_pair_name`     | Name of the created key pair                |
| `control_instance`  | Details of the deployed EC2 instance        |

## Resources

- `aws_key_pair`: Creates an SSH key pair for the instance.
- `aws_instance`: Provisions the EC2 control instance.
- `aws_eip_association`: Associates the Elastic IP with the instance.

## Example Usage

```hcl
module "controller" {
    source                  = "./controller"
    public_key_path         = "path/to/public_key.pub"
    ami                     = "ami-xxxxxxxx"
    security_group_ids      = ["sg-xxxxxxxx"]
    control_instance_eip_id = "eipalloc-xxxxxxxx"
    vpc_cidr_block          = "10.0.0.0/16"
    subnet_id               = "subnet-xxxxxxxx"
}
```

## Notes

- Ensure the public key file exists and is accessible.
- The instance type is set to `t3.nano` by default.
- Security groups must allow necessary traffic (e.g., SSH).
