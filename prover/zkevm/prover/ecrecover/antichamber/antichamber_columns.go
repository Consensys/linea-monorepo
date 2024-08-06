package antichamber

import "github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"

func (ac *Antichamber) cols(withActive bool) []ifaces.Column {
	res := []ifaces.Column{
		ac.ID,
		ac.IsPushing,
		ac.IsFetching,
		ac.Source,
	}
	if withActive {
		res = append(res, ac.IsActive)
	}
	return res
}

func (ec *EcRecover) cols() []ifaces.Column {
	return []ifaces.Column{
		ec.EcRecoverID,
		ec.Limb,
		ec.SuccessBit,
		ec.EcRecoverIndex,
		ec.EcRecoverIsData,
		ec.EcRecoverIsRes,
	}
}

func (ad *Addresses) cols() []ifaces.Column {
	return []ifaces.Column{
		ad.addressHiUntrimmed,
		ad.addressHi,
		ad.addressLo,
	}
}

func (ts *txSignature) cols() []ifaces.Column {
	return []ifaces.Column{
		ts.isTxHash,
		ts.txHashHi,
		ts.txHashLo,
	}
}

func (ugd *UnalignedGnarkData) cols() []ifaces.Column {
	return []ifaces.Column{
		ugd.IsPublicKey,
		ugd.GnarkIndex,
		ugd.GnarkData,
	}
}

func (ac *Antichamber) unalignedGnarkDataSource() *unalignedGnarkDataSource {
	return &unalignedGnarkDataSource{
		IsActive:   ac.IsActive,
		IsPushing:  ac.IsPushing,
		IsFetching: ac.IsFetching,
		Source:     ac.Source,
		Limb:       ac.EcRecover.Limb,
		SuccessBit: ac.EcRecover.SuccessBit,
		IsData:     ac.EcRecover.EcRecoverIsData,
		IsRes:      ac.EcRecover.EcRecoverIsRes,
		TxHashHi:   ac.txSignature.txHashHi,
		TxHashLo:   ac.txSignature.txHashLo,
	}
}
