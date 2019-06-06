package metrics

type nodeStatus struct {
	up      bool
	down    bool
	joining bool
	leaving bool
	moving  bool
}

type labelPair struct {
	liveness string
	state    string
}

func (n *nodeStatus) normal() bool {
	return !n.joining && !n.leaving && !n.moving
}

func (n *nodeStatus) livenessLabel() string {
	if n.up {
		return "up"
	}
	return "down"
}

func (n *nodeStatus) stateLabel() string {
	if n.joining {
		return "joining"
	} else if n.leaving {
		return "leaving"
	} else if n.moving {
		return "moving"
	}
	return "normal"
}

var allLabelPairs = []labelPair{
	{"up", "normal"},
	{"up", "joining"},
	{"up", "leaving"},
	{"up", "moving"},
	{"down", "normal"},
	{"down", "joining"},
	{"down", "leaving"},
	{"down", "moving"},
}

func (n *nodeStatus) unapplicableLabelPairs() []labelPair {
	applicableLabelPair := labelPair{n.livenessLabel(), n.stateLabel()}
	var unapplicableLabelPairs []labelPair

	for _, l := range allLabelPairs {
		if l != applicableLabelPair {
			unapplicableLabelPairs = append(unapplicableLabelPairs, l)
		}
	}

	return unapplicableLabelPairs
}

func transformClusterStatus(clusterStatus *clusterStatus) map[string]*nodeStatus {
	podIPToNodeStatus := make(map[string]*nodeStatus)

	for _, liveNode := range clusterStatus.liveNodes {
		if _, ok := podIPToNodeStatus[liveNode]; !ok {
			podIPToNodeStatus[liveNode] = &nodeStatus{}
		}
		podIPToNodeStatus[liveNode].up = true
		podIPToNodeStatus[liveNode].down = false
	}

	for _, unreachableNode := range clusterStatus.unreachableNodes {
		if _, ok := podIPToNodeStatus[unreachableNode]; !ok {
			podIPToNodeStatus[unreachableNode] = &nodeStatus{}
		}
		podIPToNodeStatus[unreachableNode].up = false
		podIPToNodeStatus[unreachableNode].down = true
	}

	for _, leavingNode := range clusterStatus.leavingNodes {
		if _, ok := podIPToNodeStatus[leavingNode]; !ok {
			podIPToNodeStatus[leavingNode] = &nodeStatus{}
		}
		podIPToNodeStatus[leavingNode].leaving = true
		podIPToNodeStatus[leavingNode].joining = false
		podIPToNodeStatus[leavingNode].moving = false
	}

	for _, movingNode := range clusterStatus.movingNodes {
		if _, ok := podIPToNodeStatus[movingNode]; !ok {
			podIPToNodeStatus[movingNode] = &nodeStatus{}
		}
		podIPToNodeStatus[movingNode].moving = true
		podIPToNodeStatus[movingNode].joining = false
		podIPToNodeStatus[movingNode].leaving = false
	}

	for _, joiningNode := range clusterStatus.joiningNodes {
		if _, ok := podIPToNodeStatus[joiningNode]; !ok {
			podIPToNodeStatus[joiningNode] = &nodeStatus{}
		}
		podIPToNodeStatus[joiningNode].joining = true
		podIPToNodeStatus[joiningNode].moving = false
		podIPToNodeStatus[joiningNode].leaving = false
	}

	return podIPToNodeStatus
}

// IsUpAndNormal checks that this node is UP and NORMAL
func (n *nodeStatus) IsUpAndNormal() bool {
	return n.up && n.normal()
}
