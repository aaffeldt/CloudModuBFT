package peer

import "fmt"

func (p Node) logRoleInitialization() {
	switch p.myRole {
	case ROLE_CLIENT:
		fmt.Printf("Initialized as CLIENT. Id: %d\n", p.myID)
	case ROLE_PEER_FOLLOWER:
		fmt.Printf("Initialized as FOLLOWER. Id: %d\n", p.myID)
	case ROLE_PEER_LEADER:
		fmt.Printf("Initialized as LEADER. Id: %d\n", p.myID)
	case ROLE_COORDINATOR:
		fmt.Printf("Initialized as COORDINATOR. Id: %d\n", p.myID)
	}
}
