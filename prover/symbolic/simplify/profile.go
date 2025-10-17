package simplify

import (
	"math"
	"math/rand/v2"

	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// profiles an expression
func ProfileExpression(e *symbolic.Expression) {

	board := e.Board()
	rng := rand.New(utils.NewRandSource(42))

	// This build up the marginal computation cost for every node. This is
	// computed by computing the cost of every children and adding the overhead
	// of computing the current node. The resulting cost is then divided by the
	// number of parent to have multiple count of the same node.
	nodeCosts := make([][]float64, len(board.Nodes))
	for level := 1; level < len(board.Nodes); level++ {
		nodeCosts[level] = make([]float64, len(board.Nodes[level]))
		for i := range nodeCosts[level] {
			costStats := evaluateNodeCosts(board.Nodes[level][i])
			nodeCosts[level][i] = float64(costStats.NumAdd) + float64(costStats.NumMul)
			childrenID := board.Nodes[level][i].Children

			for _, childID := range childrenID {
				childLevel, childPos := childID.Level(), childID.PosInLevel()

				// The child is a leaf and therefore its cost is zero
				if childLevel == 0 {
					continue
				}

				childCost := nodeCosts[childLevel][childPos]
				nodeCosts[level][i] += childCost / float64(len(board.Nodes[childLevel][childPos].Parents))
			}
		}
	}

	// This randomly picks one leaf by descending the tree randomly and favoring
	// the subtrees with the highest cost.
	pickLeaf := func() string {

		var (
			currLevel = len(board.Nodes) - 1
			currPos   = 0
		)

		for currLevel > 0 {

			var (
				node         = board.Nodes[currLevel][currPos]
				childrenCost = make([]float64, len(node.Children))
			)

			for i, childID := range node.Children {
				childLevel, childPos := childID.Level(), childID.PosInLevel()
				// The child is a leaf and therefore its intrinsic cost is zero
				if childLevel == 0 {
					continue
				}
				childrenCost[i] = nodeCosts[childLevel][childPos]
			}

			switch e := node.Operator.(type) {
			case symbolic.Product:
				for i := range node.Children {
					childrenCost[i] += math.Abs(float64(e.Exponents[i]))
				}
			case symbolic.LinComb:
				for i := range node.Children {
					childrenCost[i] += 1.
					if math.Abs(float64(e.Coeffs[i])) > 1. {
						childrenCost[i] += 1.
					}
				}
			case symbolic.PolyEval:
				for i := range node.Children {
					if i == 0 {
						childrenCost[i] += float64(len(node.Children) - 2)
						continue
					}
					childrenCost[i] += 1.
				}
			default:
				utils.Panic("unexpected type %T", e)
			}

			picked := utils.RandChooseWeighted(rng, childrenCost)
			currLevel, currPos = node.Children[picked].Level(), node.Children[picked].PosInLevel()
		}

		finalLeaf := board.Nodes[0][currPos]

		switch e := finalLeaf.Operator.(type) {
		case symbolic.Constant:
			return "<constant>"
		case symbolic.Variable:
			return e.Metadata.String()
		default:
			panic("unexpected")
		}
	}

	chosedLeaves := []string{}
	numSamples := 10_000
	for i := 0; i < numSamples; i++ {
		leaf := pickLeaf()
		chosedLeaves = append(chosedLeaves, leaf)
		logrus.Infof("sample %v: %v", i, leaf)
	}
}
