package mesh

const (
	NODE_OK            = 0
	NODE_RETRY         = 1
	NODE_TIMEOUT       = 2
	NODE_TIMEOUT_RETRY = 3
	NODE_DEAD          = 4 //Never used - Node will be removed after timeout retry routine
)
