# Spread Placement Groups Across Availability Zones (Terraform)

This Terraform configuration provisions AWS EC2 instances distributed across multiple availability zones using spread placement groups. Spread placement groups ensure that each instance is placed on distinct hardware, reducing the risk of simultaneous failures.

## Purpose

- **High Availability:** Instances are distributed across racks and availability zones to minimize correlated failures.
- **Customizable Topology:** Supports configurable numbers of nodes, clients, and placement groups.
- **Modular Design:** Uses modules for VPC, controller, nodes, and clients.

## Main Resources

- **aws_placement_group:** Creates spread placement groups for distributing instances.
- **modules:** Provisions VPC, controller, nodes, and clients with appropriate networking and security.

## Variables

| Name                        | Description                                         | Type         | Default / Required |
|-----------------------------|-----------------------------------------------------|--------------|--------------------|
| `region`                    | AWS region to deploy resources                      | string       | `eu-central-1`     |
| `ami`                       | AMI ID for EC2 instances                            | string       | `ami-0a116fa7c861dd5f9` |
| `instance_type`             | Instance type for nodes                             | string       | `t3.nano`          |
| `client_instance_type`      | Instance type for clients                           | string       | required           |
| `number_of_nodes`           | Number of node instances                            | number       | required           |
| `number_of_clients`         | Number of client instances                          | number       | required           |
| `number_of_node_spread_groups` | Number of node spread placement groups           | number       | required           |
| `list_of_node_azs`          | List of AZs for nodes                               | list(string) | required           |
| `list_of_client_azs`        | List of AZs for clients                             | list(string) | required           |
| `list_of_node_spread_group` | Spread group index for each node                    | list(number) | required           |
| `list_of_client_spread_group` | Spread group index for each client                | list(number) | required           |
| `public_key_path`           | Path to SSH public key for controller               | string       | required           |

## Usage

1. **Configure Variables:** Edit variables in `main.tf` or use a `terraform.tfvars` file.
2. **Initialize Terraform:**  
   ```
   terraform init
   ```
3. **Apply Configuration:**  
   ```
   terraform apply
   ```
4. **Destroy Resources:**  
   ```
   terraform destroy
   ```

## Notes

- Each spread placement group supports a maximum of 7 running instances per AZ.
- If more than 7 instances per AZ are needed, use multiple spread placement groups.
- Placement groups are not supported for Dedicated Instances.

## References

- [AWS EC2 Placement Groups Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html)
