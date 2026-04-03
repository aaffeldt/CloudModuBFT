
import sys 


def process(list_of_cluster,list_of_roles, max_partitions):
    
    partition_count = {}
    client_partition_index = 0
    partition = []
    
    for r,c in zip(list_of_roles,list_of_cluster):
        if c not in partition_count:
            partition_count[c] = 0
        
        if r == 0: 
            partition.append(partition[client_partition_index])
            client_partition_index += 1
        else:
            partition.append((partition_count[c] % max_partitions)+1)
            partition_count[c] += 1
        
    return ",".join(map(str,partition))
   
   
    
    
    
    
            

        
    
    

def main():
    if len(sys.argv) != 4:
        print("Usage: python generate-list-of-buckets.py <argument>")
        sys.exit(1)
    
    list_of_cluster = sys.argv[1].split(",")
    list_of_roles = map(int,sys.argv[2].split(","))
    max_partitions = int(sys.argv[3])
    
    
    
    
    
    
    
    
    # print(f"Received argument: {argument}")
    print(process(list_of_cluster,list_of_roles, max_partitions))


if __name__ == "__main__":
    main()