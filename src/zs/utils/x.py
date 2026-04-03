import sys

# azs_node="a,b,c,a,a,c,b,c,a"
def process(azs_node:str):
    azs=azs_node.split(",")
    distinct_azs=sorted(list(set(azs)))
    
    counts={}
    for daz in distinct_azs:
        counts[daz] = 0 
    s=""
    for az in azs:
        s+=f"{counts[az]},"
        counts[az]+=1
        
    print(s[:-1])
        
        
    
    

def main():
    if len(sys.argv) != 2:
        print("Usage: python x.py <argument>")
        sys.exit(1)
    
    argument = sys.argv[1]
    # print(f"Received argument: {argument}")
    process(argument)


if __name__ == "__main__":
    main()