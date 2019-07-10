package prover

import (
	"github.com/spacemeshos/merkle-tree"
	"github.com/spacemeshos/merkle-tree/cache"
	"github.com/spacemeshos/poet/shared"
)

const MerkleMinCacheLayer = 0  // Merkle nodes from this layer up will be cached, in addition to the base layer
const MerkleMinMemoryLayer = 2 // Below this layer caching is done on-disk, from this layer up -- in-memory

// GetProof computes the PoET DAG, uses Fiat-Shamir to derive a challenge from the Merkle root and generates a Merkle
// proof using the challenge and the DAG.
func GetProof(labelHashFunc func(data []byte) []byte, merkleHashFunc func(lChild, rChild []byte) []byte,
	leafCount uint64, securityParam uint8) (shared.MerkleProof, error) {

	metaFactory := NewReadWriterMetaFactory(MerkleMinMemoryLayer)
	defer metaFactory.Cleanup()
	treeCache := cache.NewWriter(
		cache.Combine(
			cache.SpecificLayersPolicy(map[uint]bool{0: true}),
			cache.MinHeightPolicy(MerkleMinCacheLayer)),
		metaFactory.GetFactory())
	tree, err := merkle.NewTreeBuilder().WithHashFunc(merkleHashFunc).WithCacheWriter(treeCache).Build()
	if err != nil {
		return shared.MerkleProof{}, err
	}

	for leafID := uint64(0); leafID < leafCount; leafID++ {
		err := tree.AddLeaf(shared.MakeLabel(labelHashFunc, leafID, tree.GetParkedNodes()))
		if err != nil {
			return shared.MerkleProof{}, err
		}
	}
	root := tree.Root()

	cacheReader, err := treeCache.GetReader()
	if err != nil {
		return shared.MerkleProof{}, err
	}
	provenLeafIndices := shared.FiatShamir(root, leafCount, securityParam)
	_, provenLeaves, proofNodes, err := merkle.GenerateProof(provenLeafIndices, cacheReader)
	if err != nil {
		return shared.MerkleProof{}, err
	}

	return shared.MerkleProof{
		Root:         root,
		ProvenLeaves: provenLeaves,
		ProofNodes:   proofNodes,
	}, nil
}