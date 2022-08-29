package mesh

const (
	NODE_OK            = 1
	NODE_RETRY         = 2
	NODE_TIMEOUT       = 3
	NODE_TIMEOUT_RETRY = 4
	NODE_DEAD          = 5 //Never used - Node will be removed after timeout retry routine
)
