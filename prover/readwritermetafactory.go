package prover

import (
	"fmt"
	"github.com/spacemeshos/merkle-tree/cache"
	"github.com/spacemeshos/merkle-tree/cache/readwriters"
	"github.com/spacemeshos/smutil/log"
	"os"
)

// ReadWriterMetaFactory generates Merkle LayerFactory functions. The functions it creates generate file read-writers
// starting from the base layer and up to minMemoryLayer-1. From minMemoryLayer and up the functions generate slice
// read-writers.
// The MetaFactory tracks the files it creates and removes them when Cleanup() is called.
type ReadWriterMetaFactory struct {
	minMemoryLayer uint
	filesCreated   map[string]bool
}

// NewReadWriterMetaFactory returns a new ReadWriterMetaFactory.
func NewReadWriterMetaFactory(minMemoryLayer uint) *ReadWriterMetaFactory {
	return &ReadWriterMetaFactory{
		minMemoryLayer: minMemoryLayer,
		filesCreated:   make(map[string]bool),
	}
}

// GetFactory creates a Merkle LayerFactory function.
func (mf *ReadWriterMetaFactory) GetFactory() cache.LayerFactory {
	return func(layerHeight uint) (cache.LayerReadWriter, error) {
		if layerHeight < mf.minMemoryLayer {
			fileName := makeFileName(layerHeight)
			readWriter, err := readwriters.NewFileReadWriter(fileName)
			if err != nil {
				return nil, err
			}
			mf.filesCreated[fileName] = true
			return readWriter, nil
		}
		return &readwriters.SliceReadWriter{}, nil
	}
}

// Cleanup removes the files that were created by the LayerFactory functions generated by this MetaFactory.
func (mf *ReadWriterMetaFactory) Cleanup() {
	failedRemovals := make(map[string]bool)
	for filename := range mf.filesCreated {
		err := os.Remove(filename)
		if err != nil {
			log.Error("could not remove temp file %v: %v", filename, err)
			failedRemovals[filename] = true
		}
	}
	mf.filesCreated = failedRemovals
}

func makeFileName(layer uint) string {
	return fmt.Sprintf("poet_layercache_%d.bin", layer)
}