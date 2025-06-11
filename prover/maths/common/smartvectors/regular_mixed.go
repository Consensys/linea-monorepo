package smartvectors

import "github.com/consensys/linea-monorepo/prover/maths/common/mempool"

func AllocFromPoolMixed(pool mempool.MemPool) *PooledExt {
	poolPtr := pool.AllocExt()
	return &PooledExt{
		RegularExt: *poolPtr,
		poolPtr:    poolPtr,
	}
}
