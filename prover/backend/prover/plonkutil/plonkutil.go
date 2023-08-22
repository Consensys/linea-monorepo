package plonkutil

import (
	"bufio"
	"bytes"
	"io"
	"path"
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/backend/plonk"
	plonkBn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

// Serializes the proof in an 0x prefixed hexstring
func SerializeProof(proof plonk.Proof) string {
	proof_ := proof.(*plonkBn254.Proof)
	proofBytes := proof_.MarshalSolidity()
	var buf bytes.Buffer
	proof.WriteRawTo(&buf)
	return hexutil.Encode(proofBytes)
}

// Export the setup
func ExportToFile(pp Setup, folderOut string) {

	solidityPath := path.Join(folderOut, "verifierContract.sol")
	f := files.MustOverwrite(solidityPath)

	err := pp.VK.ExportSolidity(f)
	if err != nil {
		panic(err)
	}

	vkFile := files.MustOverwrite(path.Join(folderOut, "verifying_key.bin"))
	_, err = pp.VK.WriteTo(vkFile)
	if err != nil {
		panic(err)
	}
	vkFile.Close()

	pkFile := files.MustOverwrite(path.Join(folderOut, "proving_key.bin"))
	buf := bufio.NewWriterSize(pkFile, 50_000_000_000)
	_, err = pp.PK.(*plonkBn254.ProvingKey).WriteRawTo(buf)
	if err != nil {
		panic(err)
	}
	pkFile.Close()

	csFile := files.MustOverwrite(path.Join(folderOut, "circuit.bin"))
	_, err = pp.SCS.WriteTo(csFile)
	if err != nil {
		panic(err)
	}
	csFile.Close()
}

// Public parameters for the test circuit in Plonk
type Setup struct {
	PK  plonk.ProvingKey
	VK  plonk.VerifyingKey
	SCS constraint.ConstraintSystem
}

// Read setup from the config object
func ReadPPFromConfig(setup chan Setup) {
	conf := config.MustGetProver()
	pp := Setup{
		PK:  &plonkBn254.ProvingKey{},
		VK:  &plonkBn254.VerifyingKey{},
		SCS: &cs.SparseR1CS{},
	}

	wg := new(sync.WaitGroup)
	wg.Add(3)

	// read the proving key
	go ReadProvingKey(conf.PKeyFile, pp.PK, wg)
	go readWithStats("r1cs", conf.R1CSFile, pp.SCS, wg)
	go readWithStats("verifying-key", conf.VKeyFile, pp.VK, wg)

	wg.Wait()
	logrus.Infof("parsed r1cs with %v instructions, %v constraints", pp.SCS.GetNbInstructions(), pp.SCS.GetNbConstraints())
	setup <- pp
}

func ReadProvingKey(fromFile string, pk plonk.ProvingKey, wg *sync.WaitGroup) {
	defer wg.Done()
	pk_ := pk.(*plonkBn254.ProvingKey)

	logrus.Infof("Reading proving-key from %q", fromFile)
	f := files.MustRead(fromFile)
	defer f.Close()
	logrus.Infof("Successfully read proving-key from %q", fromFile)

	stats, err := f.Stat()
	if err != nil {
		utils.Panic("error reading proving-key : %v", err)
	}
	logrus.Infof("The file proving-key has size %d bytes", stats.Size())

	logrus.Infof("Bufferize the proving key")
	reader := bufio.NewReaderSize(f, int(stats.Size()))

	n, err := pk_.UnsafeReadFrom(reader)
	if err != nil {
		utils.Panic("error reading proving-key : %v", err)
	}

	logrus.Infof("Successfully parsed %v bytes of proving-key", n)

}

func readWithStats(name string, fromFile string, into io.ReaderFrom, wg *sync.WaitGroup) {
	defer wg.Done()
	logrus.Infof("Reading %q from %q", name, fromFile)
	f := files.MustRead(fromFile)
	defer f.Close()
	logrus.Infof("Successfully read %q from %q", name, fromFile)

	stats, err := f.Stat()
	if err != nil {
		utils.Panic("error reading %v : %v", name, err)
	}

	logrus.Infof("The file %q has size %d bytes", name, stats.Size())
	n, err := into.ReadFrom(f)
	if err != nil {
		utils.Panic("error reading %v : %v", name, err)
	}

	logrus.Infof("Successfully parsed %v bytes of %v", n, name)
}
