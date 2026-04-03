package peer

import "compress/gzip"

const (
	MAX_SIG_VERIF        = 10 //6
	MAX_SIG_CREATE       = 8  // 8 there's 3x this number of threads in verification
	MAX_SIG_VERIF_CLIENT = 20 //20
	MAXBUF               = 2048 * 2048
	CHECKPOINT_INTERVAL  = 500
	CIRCULAR_BUFFER_SIZE = 2000
	LOG_SIZE             = 10000
)
const (
	ROLE_CLIENT        = 0
	ROLE_PEER_FOLLOWER = 1
	ROLE_PEER_LEADER   = 2
	ROLE_COORDINATOR   = 3
)

const (
	PORT_CLIENT = 12340
	PORT_PEER   = 12341
	PORT_COORD  = 12342
)

const ALLOWED_PARALLEL_CLUSTER_PREPARES = 2

const (
	COMPRESSION = gzip.DefaultCompression
)
