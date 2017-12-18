package network

type QuorumSlice struct {
	// Members is a list of public keys for nodes that occur in the quorum slice.
	// Members must be unique.
	// Typically includes ourselves.
	Members []string

	// The number of members we require for consensus, including ourselves.
	// The protocol can support other sorts of slices, like weighted or any wacky
	// thing, but for now we only do this simple "any k out of these n" voting.
	Threshold int
}

func (qs *QuorumSlice) atLeast(nodes []string, t int) bool {
	count := 0
	for _, member := range qs.Members {
		for _, node := range nodes {
			if node == member {
				count++
				if count >= t {
					return true
				}
				break
			}
		}
	}
	return false
}

func (qs *QuorumSlice) BlockedBy(nodes []string) bool {
	return qs.atLeast(nodes, len(qs.Members) - qs.Threshold + 1)
}

func (qs *QuorumSlice) SatisfiedWith(nodes []string) bool {
	return qs.atLeast(nodes, qs.Threshold)
}

type QuorumFinder interface {
	QuorumSlice(node string) (*QuorumSlice, bool)
	PublicKey() string
}

// Returns whether this set of nodes meets the quorum for the network overall.
func MeetsQuorum(f QuorumFinder, nodes []string) bool {
	// Filter out the nodes in the potential quorum that do not have their
	// own quorum slices met
	hasUs := false
	filtered := []string{}
	for _, node := range nodes {
		qs, ok := f.QuorumSlice(node)
		if ok && qs.SatisfiedWith(nodes) {
			filtered = append(filtered, node)
			if node == f.PublicKey() {
				hasUs = true
			}
		}
	}
	if !hasUs {
		return false
	}
	if len(filtered) == len(nodes) {
		return true
	}
	return MeetsQuorum(f, filtered)
}
