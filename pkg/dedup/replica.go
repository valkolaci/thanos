package dedup

import (
	"sort"

	"github.com/prometheus/prometheus/storage"
	"github.com/thanos-io/thanos/pkg/store/storepb"
)

// replicaAwareSortSet is a set that re-sorts the input set series in order to deduplicate replicas.
// TODO(bwplotka): Consider algorithm that uses the fact that input series are sorted, do not require full buffering.
type replicaAwareSortSet struct {
	set           storage.SeriesSet
	replicaLabels map[string]struct{}

	initialized bool
	i           int
	buff        []storage.Series
}

func newReplicaAwareSortSet(set storage.SeriesSet, replicaLabels map[string]struct{}) *replicaAwareSortSet {
	return &replicaAwareSortSet{
		set:           set,
		replicaLabels: replicaLabels,
		buff:          make([]storage.Series, 0, 1024),
	}
}
func (r *replicaAwareSortSet) Next() bool {
	if !r.initialized {
		for r.set.Next() {
			r.buff = append(r.buff, r.set.At())
		}
		r.initialized = true

		cmpFunc := storepb.NewReplicaAwareLabelsCompareFunc(r.replicaLabels)
		sort.Slice(r.buff, func(i, j int) bool {
			return cmpFunc(r.buff[i].Labels(), r.buff[j].Labels()) < 0
		})
		r.i = -1
	}

	if r.set.Err() != nil || r.i >= len(r.buff)-1 {
		return false
	}
	r.i++
	return true
}

func (r *replicaAwareSortSet) At() storage.Series {
	return r.buff[r.i]
}

func (r *replicaAwareSortSet) Err() error {
	return r.set.Err()
}

func (r *replicaAwareSortSet) Warnings() storage.Warnings {
	return r.Warnings()
}
