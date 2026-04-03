# Utils (scripts) — CloudModuBFT

Purpose
- Small collection of helper scripts used when generating node and cluster assignments for deployments.
- Scripts produce comma-separated lists consumed by Terraform variables or other tooling in this repo.

Contents
- generate-list-of-roles.sh
  - Purpose: produce a comma-separated list of numeric "role/partition" values for every node in deployment order.
  - Arguments:
    1. number_of_leaders
    2. number_of_leader_follower (followers per leader)
    3. number_of_clients
    4. number_of_partitions
  - Output: single line, comma-separated numbers (no spaces).
    - The script appends values in this order: leaders, clients, leader-followers.
    - Note: partitions in the output are 1-based (values range from 1..number_of_partitions).
  - Example:
    - ./generate-list-of-roles.sh 3 2 2 3
    - Output (example): 1,2,3,1,2,1,2,2
  - Notes:
    - Minimal / no argument validation — ensure you pass correct integer args.
    - Make executable: chmod +x generate-list-of-roles.sh
    - Shebang present: script runs with bash.

- generate-list-of-cluster.sh
  - Purpose: produce a comma-separated list of cluster indices for every node in deployment order.
  - Arguments:
    1. number_of_leaders
    2. number_of_leader_follower (followers per leader)
    3. number_of_clients
  - Output: single line, comma-separated cluster indices (0-based).
    - Order: leaders, clients, leader-followers.
  - Example:
    - ./generate-list-of-cluster.sh 3 2 2
    - Output (example): 0,1,2,0,1,0,1,2
  - Notes:
    - Cluster indices are 0-based (leaders are assigned clusters 0..N-1).
    - Make executable: chmod +x generate-list-of-cluster.sh

Usage tips
- Consume outputs directly in shell or export to variables:
  - roles=$(./generate-list-of-roles.sh 3 2 2 3)
  - clusters=$(./generate-list-of-cluster.sh 3 2 2)
- Use produced CSV strings to populate Terraform variable files or pass to modules expecting comma-separated values.
- Validate counts and ranges before using in production; these scripts are lightweight helpers intended for simple workflows.

Troubleshooting & improvements
- Add argument validation to ensure numeric and sensible ranges.
- Consider producing JSON arrays for easier parsing in downstream tools.
- If you change numbering conventions (0-based vs 1-based), update callers and Terraform expectations accordingly.

Location
- Scripts live under:
  src/aws/utils/
- Related infrastructure modules that consume these outputs:
  src/aws/infrastructure/modules/ (node, vpc, controller, etc.)
