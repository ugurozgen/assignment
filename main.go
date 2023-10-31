package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/multi"
	"gonum.org/v1/gonum/graph/path"
)

// Define the available pack sizes
var packSizes = []int{250, 500, 1000, 2000, 5000}

// RequiredPacks is a map of pack sizes and the number required
type RequiredPacks map[int]int

// multigraph of quantities, allowing for multiple weights (lines) between two nodes (edge)
type quantityGraph struct {
	packSizeCount int
	candidates    map[int]quantityNode
	*multi.WeightedDirectedGraph
}

type quantityNode struct {
	quantity int
}

// Calculate the minimum number of packs required to fulfill the order
func calculatePacks(quantity int) (RequiredPacks, error) {
	packs := make(RequiredPacks)

	if quantity <= 0 {
		return packs, nil
	}

	// create a graph with the initial quantity as root node
	qGraph := quantityGraph{
		packSizeCount:         len(packSizes),
		candidates:            make(map[int]quantityNode),
		WeightedDirectedGraph: multi.NewWeightedDirectedGraph(),
	}
	rootNode := quantityNode{quantity}
	qGraph.AddNode(rootNode)

	// generate permutations by recursively subtracting packs, with pack sizes cascading in descending order
	// e.g. [5 4 3 2 1] -> [4 3 2 1] -> [3 2 1] -> [2 1] -> [1]
	for i := len(packSizes); i >= 1; i-- {
		availableSizes := make([]int, i)
		copy(availableSizes, packSizes[:i])
		sort.Sort(sort.Reverse(sort.IntSlice(availableSizes)))
		qGraph.subtractPacks(rootNode, availableSizes)
	}

	// aid traversal by removing unnecessary nodes
	candidateNode := qGraph.closestCandidate()
	qGraph.pruneNodes(candidateNode)

	// find the shortest path to the quantity closest to zero
	shortest, _ := path.AStar(rootNode, candidateNode, qGraph, nil)
	path, _ := shortest.To(candidateNode.ID())
	pathLength := len(path)

	// count each weighted line that forms the path as a used pack size
	for i, currentNode := range path {
		nextIndex := i + 1
		if nextIndex >= pathLength {
			break
		}

		lines := qGraph.WeightedDirectedGraph.WeightedLines(currentNode.ID(), path[nextIndex].ID())
		lines.Next()
		packs[int(lines.WeightedLine().Weight())]++
	}

	return packs, nil
}

func (g *quantityGraph) subtractPacks(n quantityNode, packSizes []int) {
	// stop generating permutations if we've found more paths to 0 than available pack sizes
	if nodesToZero := g.To(int64(0)); nodesToZero.Len() >= g.packSizeCount {
		return
	}

	for _, size := range packSizes {
		// find or create a node by the subtracted quantity
		nextQuantity := n.quantity - size
		nextNode := quantityNode{nextQuantity}
		if existingNode := g.Node(nextNode.ID()); existingNode == nil {
			g.AddNode(nextNode)
		}

		// maintain unique weights for edges between two quantities to avoid unnecessary recalculations
		weight := float64(size)
		if g.hasWeightedLineFromTo(n, nextNode, weight) {
			continue
		}

		// link the nodes by pack size
		g.SetWeightedLine(g.NewWeightedLine(n, nextNode, weight))

		// track nodes which satisfy the required quantity, stopping at this depth
		if nextQuantity <= 0 {
			g.candidates[nextQuantity] = nextNode
			continue
		}

		// subtract from the next quantity, increasing depth
		g.subtractPacks(nextNode, packSizes)
	}
}

func (g *quantityGraph) hasWeightedLineFromTo(from graph.Node, to graph.Node, weight float64) bool {
	for _, line := range graph.WeightedLinesOf(g.WeightedLines(from.ID(), to.ID())) {
		if line.Weight() == weight {
			return true
		}
	}
	return false
}

func (g quantityGraph) Weight(xid, yid int64) (w float64, ok bool) {
	return path.UniformCost(g)(xid, yid)
}

func (g *quantityGraph) closestCandidate() quantityNode {
	// create a slice of quantities from the map keys
	quantities := make([]int, len(g.candidates))
	i := 0
	for k := range g.candidates {
		quantities[i] = k
		i++
	}

	// reverse sort so the closest candidate is first
	sort.Sort(sort.Reverse(sort.IntSlice(quantities)))

	return g.candidates[quantities[0]]
}

func (g *quantityGraph) pruneNodes(candidate graph.Node) {
	// remove other candidates from the graph
	for _, node := range g.candidates {
		if node != candidate {
			g.RemoveNode(node.ID())
		}
	}

	// remove nodes which don't have any edges going out
	var retraverse bool
	for {
		retraverse = false
		it := g.Nodes()
		for it.Next() {
			if node := it.Node(); node != candidate && len(graph.NodesOf(g.From(node.ID()))) == 0 {
				g.RemoveNode(node.ID())
				retraverse = true
			}
		}
		if !retraverse {
			break
		}
	}
}

func (n quantityNode) ID() int64 {
	return int64(n.quantity)
}

func calculatePackHandler(c *gin.Context) {
	itemCountStr := c.Param("itemCount")
	itemCount, _ := strconv.Atoi(itemCountStr)

	packs, _ := calculatePacks(itemCount)

	c.JSON(http.StatusOK, gin.H{"itemCount": itemCount, "packs": packs})
}

func main() {
	router := gin.Default()
	router.GET("/order/:itemCount", calculatePackHandler)
	router.Run(":8080")
}
