// @task spec/tasks/mac-storage-scout.task-03.md
// @purpose Aggregate scan events into threshold-aware tree.
package aggregate

import (
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"path/filepath"
	"sort"
)

// MssTreeAggregatorAdapter builds deterministic threshold-aware trees from walk events.
//
// @purpose Build deterministic threshold-aware trees for report rendering.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @implements {MssNodeAggregatorPort} internal/mss/ports/mss_node_aggregator_port.go
type MssTreeAggregatorAdapter struct{}

// @see {MssNodeAggregatorPort#BuildTree} internal/mss/ports/mss_node_aggregator_port.go
// @purpose Aggregate events into deterministic threshold-aware roots.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @pre cfg.TopN >= 1 and cfg.ThresholdBytes > 0.
// @param events Walk events from filesystem walker.
// @param cfg Scan configuration.
// @returns Aggregated roots and optional validation/aggregation error.
// @post Every returned root preserves explicit-vs-other threshold contract.
func (a *MssTreeAggregatorAdapter) BuildTree(events []domain.MssWalkEvent, cfg domain.MssScanConfig) ([]*domain.MssNode, error) {
	if cfg.TopN < 1 {
		return nil, fmt.Errorf("[MssTreeAggregatorAdapter.BuildTree] topN must be >= 1")
	}
	if cfg.ThresholdBytes <= 0 {
		return nil, fmt.Errorf("[MssTreeAggregatorAdapter.BuildTree] threshold must be positive")
	}

	// START_INDEX_NODES_FROM_EVENTS
	// invariant: each unique path maps to exactly one node pointer in nodeByPath.
	nodeByPath := map[string]*domain.MssNode{}
	childByParent := map[string][]*domain.MssNode{}

	ensureNode := func(path, name string, kind domain.MssEntryKind, ext string) *domain.MssNode {
		if n, ok := nodeByPath[path]; ok {
			return n
		}
		n := &domain.MssNode{Path: path, Name: name, Kind: kind, Ext: ext}
		nodeByPath[path] = n
		return n
	}

	for _, ev := range events {
		if ev.Entry == nil {
			continue
		}
		e := ev.Entry
		n := ensureNode(e.Path, e.Name, e.Kind, e.Ext)
		n.SizeBytes = e.SizeBytes
		if e.ParentPath != "" && e.ParentPath != e.Path {
			parent := ensureNode(e.ParentPath, filepath.Base(e.ParentPath), domain.MssEntryKindDir, "")
			childByParent[parent.Path] = append(childByParent[parent.Path], n)
		}
	}
	// END_INDEX_NODES_FROM_EVENTS

	for p, children := range childByParent {
		nodeByPath[p].Children = dedupChildren(children)
	}

	// START_SELECT_CONFIGURED_ROOTS
	rootSet := map[string]bool{}
	for _, rp := range cfg.Paths {
		ap := rp
		if !filepath.IsAbs(ap) {
			if abs, err := filepath.Abs(ap); err == nil {
				ap = abs
			}
		}
		rootSet[ap] = true
	}
	roots := []*domain.MssNode{}
	for p, n := range nodeByPath {
		if rootSet[p] {
			roots = append(roots, n)
		}
	}
	// END_SELECT_CONFIGURED_ROOTS

	// START_APPLY_AGGREGATION_POLICIES
	// purpose: ensure deterministic sizing, threshold split, and stable root ordering.
	for _, r := range roots {
		recomputeSizes(r)
		applyThreshold(r, cfg.ThresholdBytes, cfg.TopN)
	}

	sortNodes(roots)
	// END_APPLY_AGGREGATION_POLICIES
	return roots, nil
}

// dedupChildren removes duplicate nodes by path.
//
// @purpose Keep parent child list unique by node path.
// @consumer BuildTree assembly.
// @param in Child node list.
// @returns Deduplicated child node list.
func dedupChildren(in []*domain.MssNode) []*domain.MssNode {
	seen := map[string]bool{}
	out := make([]*domain.MssNode, 0, len(in))
	for _, n := range in {
		if seen[n.Path] {
			continue
		}
		seen[n.Path] = true
		out = append(out, n)
	}
	return out
}

// recomputeSizes recalculates recursive directory sizes.
//
// @purpose Recompute total size from file leaves for deterministic totals.
// @consumer BuildTree post-assembly normalization.
// @param n Root node for recursive recomputation.
// @returns Recomputed size for node.
func recomputeSizes(n *domain.MssNode) int64 {
	if n.Kind == domain.MssEntryKindFile {
		if n.SizeBytes < 0 {
			n.SizeBytes = 0
		}
		return n.SizeBytes
	}
	var total int64
	for _, c := range n.Children {
		total += recomputeSizes(c)
	}
	n.SizeBytes = total
	return total
}

// sortNodes sorts nodes by size desc and name asc.
//
// @purpose Enforce stable output order contract.
// @consumer BuildTree and threshold split flow.
// @param nodes Node list to sort in place.
func sortNodes(nodes []*domain.MssNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].SizeBytes == nodes[j].SizeBytes {
			return nodes[i].Name < nodes[j].Name
		}
		return nodes[i].SizeBytes > nodes[j].SizeBytes
	})
}

// applyThreshold splits children into explicit and other buckets.
//
// @purpose Preserve threshold contract for report generation.
// @consumer BuildTree threshold aggregation phase.
// @param n Node whose children are split.
// @param threshold Byte threshold for explicit output.
// @param topN Max visible items in other top block.
func applyThreshold(n *domain.MssNode, threshold int64, topN int) {
	if n.Kind == domain.MssEntryKindFile {
		return
	}
	for _, c := range n.Children {
		applyThreshold(c, threshold, topN)
	}
	sortNodes(n.Children)

	visible := make([]*domain.MssNode, 0, len(n.Children))
	small := make([]*domain.MssNode, 0)
	// START_SPLIT_VISIBLE_AND_OTHER
	// invariant: items >= threshold stay explicit; items < threshold move to other bucket.
	for _, c := range n.Children {
		if c.SizeBytes >= threshold {
			visible = append(visible, c)
		} else {
			small = append(small, c)
		}
	}
	// END_SPLIT_VISIBLE_AND_OTHER

	if len(small) == 0 {
		n.Children = visible
		n.Other = nil
		return
	}

	other := &domain.MssOtherBucket{Count: len(small)}
	for _, s := range small {
		other.SizeBytes += s.SizeBytes
	}

	sortNodes(small)
	limit := topN
	if limit > len(small) {
		limit = len(small)
	}
	other.TopItems = append(other.TopItems, small[:limit]...)
	for i := limit; i < len(small); i++ {
		other.RestCount++
		other.RestSize += small[i].SizeBytes
	}

	types := map[string]domain.MssExtStat{}
	collectTypesFromNodes(small, types)
	topTypes := make([]domain.MssExtStat, 0, len(types))
	for _, v := range types {
		topTypes = append(topTypes, v)
	}
	sort.SliceStable(topTypes, func(i, j int) bool {
		if topTypes[i].SizeBytes == topTypes[j].SizeBytes {
			return topTypes[i].Ext < topTypes[j].Ext
		}
		return topTypes[i].SizeBytes > topTypes[j].SizeBytes
	})
	if len(topTypes) > limit {
		for _, t := range topTypes[limit:] {
			other.TypeRestCnt += t.Count
			other.TypeRestSz += t.SizeBytes
		}
		topTypes = topTypes[:limit]
	}
	other.TypeTop = topTypes

	n.Children = visible
	n.Other = other
}

// collectTypesFromNodes aggregates extension stats from subtree.
//
// @purpose Build types section payload for other bucket.
// @consumer applyThreshold type aggregation.
// @param nodes Nodes to scan for file extensions.
// @param types Mutable extension stats accumulator.
func collectTypesFromNodes(nodes []*domain.MssNode, types map[string]domain.MssExtStat) {
	for _, n := range nodes {
		if n.Kind == domain.MssEntryKindFile {
			ext := n.Ext
			if ext == "" {
				ext = "no-ext"
			}
			cur := types[ext]
			cur.Ext = ext
			cur.Count++
			cur.SizeBytes += n.SizeBytes
			types[ext] = cur
			continue
		}
		collectTypesFromNodes(n.Children, types)
	}
}
