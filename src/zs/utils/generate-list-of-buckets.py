
import sys 


def process(list_of_cluster, list_of_roles,number_of_clients_per_bucket):

    leader_bucket_count = {}
    client_bucket_count = {}
    buckets = []
    for c,r in zip(list_of_cluster,list_of_roles): 

        if c not in leader_bucket_count:
            leader_bucket_count[c] = 0
            
        if c not in client_bucket_count:
            client_bucket_count[c] = 0
            
            
        if r == 2 : 
            buckets.append(leader_bucket_count[c])
            leader_bucket_count[c] += 1
        elif r == 1 :
            buckets.append(-1) 
        else: 
            buckets.append(client_bucket_count[c]//number_of_clients_per_bucket)
            client_bucket_count[c] += 1
            
    return ",".join(map(str,buckets))
            

        
   
   
    
    
    
    
            

        
    
    

def main():
    if len(sys.argv) != 4:
        print("Usage: python generate-list-of-buckets.py <argument>")
        sys.exit(1)
    
    list_of_cluster = list(map(int,sys.argv[1].split(",")))
    list_of_roles = list(map(int,sys.argv[2].split(",")))
    number_of_clients = int(sys.argv[3])
    
    
    
    
    
    
    # print(f"Received argument: {argument}")
    print(process(list_of_cluster, list_of_roles,number_of_clients))

if __name__ == "__main__":
    main()