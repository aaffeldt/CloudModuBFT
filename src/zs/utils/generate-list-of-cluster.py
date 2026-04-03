
import sys 


def process(number_of_cluster, number_of_buckets,number_of_peers_in_a_cluster,number_of_clients_per_bucket):
    cluster=[]
    
    for i in range(number_of_cluster*number_of_peers_in_a_cluster):
        cluster.append(i%number_of_cluster)
    
    for i in range(number_of_buckets*number_of_cluster*number_of_clients_per_bucket):
        cluster.append(i%number_of_cluster)
    
    return ",".join(map(str,cluster))
   
   
    
    
    
    
            

        
    
    

def main():
    if len(sys.argv) != 5:
        print("Usage: python generate-list-of-buckets.py <argument>")
        sys.exit(1)
    
    number_of_cluster = int(sys.argv[1])
    number_of_buckets = int(sys.argv[2])
    number_of_peers_in_a_cluster=int(sys.argv[3])
    number_of_clients = int(sys.argv[4])
    
    
    
    
    
    
    # print(f"Received argument: {argument}")
    print(process(number_of_cluster, number_of_buckets,number_of_peers_in_a_cluster,number_of_clients))


if __name__ == "__main__":
    main()