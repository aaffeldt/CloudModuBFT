
import sys 


def process(list_of_cluster, number_of_buckets,number_of_clients_per_bucket):
    
    client_count={}
    follower_count={}
    
    number_of_cluster=len(set(list_of_cluster))
    number_of_nodes=len(list_of_cluster)
    
    number_of_clients = (number_of_cluster*number_of_buckets*number_of_clients_per_bucket)
    number_of_leader = (number_of_cluster*number_of_buckets)
    
    number_of_follower_per_cluster=(number_of_nodes- number_of_clients -number_of_leader)/number_of_cluster
   
    # print(number_of_follower_per_cluster)
    
    list_of_roles=[]
    for c in reversed(list_of_cluster):
        if c not in client_count: 
            client_count[c] = 0
        if c not in follower_count: 
            follower_count[c] = 0
        
        if client_count[c] < number_of_buckets*number_of_clients_per_bucket:
            list_of_roles.append(0)
            client_count[c]+=1 
        elif follower_count[c] < number_of_follower_per_cluster:
            list_of_roles.append(1)
            follower_count[c]+=1
        else:
            list_of_roles.append(2)
        
            
        
    return ",".join(map(str,reversed(list_of_roles)))
        
            

        
    
    

def main():
    if len(sys.argv) != 4:
        print("Usage: python generate-list-of-buckets.py <argument>")
        sys.exit(1)
    
    list_of_cluster = sys.argv[1].split(",")
    number_of_buckets = int(sys.argv[2])
    number_of_clients = int(sys.argv[3])
    
    
    
    
    
    
    # print(f"Received argument: {argument}")
    print(process(list_of_cluster, number_of_buckets,number_of_clients))


if __name__ == "__main__":
    main()