// THIS FILE IS MEANT ONLY TO BE A PLACEHOLDER AND SERVE AS A MOCK DEFINING THE COMPONENTS REQUIRED
// FOR LIMITLESS PROVER. THERE ARE NO INTERACTIONS WITH THE ACTUAL CRYPTOGRAPHIC COMPONENTS.
// EACH COMPONENTS PREFIX BEGINS WITH 'M' SIGNIFYING MOCK.
// SEE https://app.diagrams.net/#G1U6S4MTrt7lsipc3TZrL4xXjvvVeghc8k#%7B%22pageId%22%3A%2206rcgNj9AqHDneUgptMC%22%7D

package mock

import "github.com/consensys/linea-monorepo/prover/backend/execution"

// Specifies the number of segments ideally set in the config. file
var segments int

// MBootStrapper initializes the prover with the necessary data
type MBootStrapper struct {
}

// MDistMetadata handles metadata about the distribution of module ID/segment ID pairs
type MDistMetadata struct {
	// Map from Module ID to Segment ID
	ModSegMap map[int]int `json:"modSegMap"`

	// Request ID
	ReqId string `json:"reqId"`
}

// MSubmoduleGLProver handles global-local proof generation
type MSubmoduleGLProver struct {
}

// MGLReq represents a request to the global-local prover
type MGLReq struct {
	ReqId                        string `json:"reqId"`
	ModuleId                     string `json:"moduleId"`
	SegmentId                    int    `json:"segmentId"`
	ConflatedExecutionTracesFile string `json:"conflatedExecutionTracesFile"`
}

// Mocked Public Inputs
type MPublicInputs struct {
}

// MGLResp represents a response from the global-local prover
type MGLResp struct {
	ModId       string          `json:"modId"`
	SegmentID   int             `json:"segmentId"`
	ModProof    string          `json:"modProof"`
	QueryResult string          `json:"queryResult"`
	Auxilliary  []MPublicInputs `json:"auxilliary"`
}

// initBootstrap initializes the bootstrapping process
// Outputs the submodule request for global-local prover for round 0
func (b MBootStrapper) initBootstrap(req execution.Request) (MGLReq, MDistMetadata, error) {
	return MGLReq{}, MDistMetadata{}, nil
}

// RandomnessBeacon provides randomness for the proof generation process
type RandomnessBeacon struct {
}

// MLPPBeaconReq represents a request for LPP beacon data
type MLPPBeaconReq struct {
	LPPColumns     []string `json:"lppColumns"`
	LPPCommitments []string `json:"lppCommitments"`
	ModuleID       string   `json:"moduleId"`
}

// generateRandomness generates randomness for the proof generation process
func (b RandomnessBeacon) generateLPPProofReq(req MLPPBeaconReq, metadata MDistMetadata) (MLPPRequest, error) {
	return MLPPRequest{}, nil
}

// proveGL generates a mock GL proof
func (gl *MSubmoduleGLProver) proveGL(req MGLReq) (MGLResp, error) {
	dummyProof := MGLResp{}
	return dummyProof, nil
}

// MSubmoduleLPPProver handles LPP proof generation
type MSubmoduleLPPProver struct {
}

// MLPPRequest represents a request for LPP proof data
type MLPPRequest struct {
	LPPReq     MLPPBeaconReq `json:"lppReq"`
	Randomseed string        `json:"randomseed"`
}

// MLPPResponse represents a response from the LPP prover
type MLPPResponse struct {
	ModuleID            string   `json:"moduleId"`
	SegmentID           string   `json:"segmentId"`
	ModuleProof         string   `json:"moduleProof"`
	QueryPartialResults []string `json:"queryPartialResults"`
}

// proveLPP generates a mock LPP proof
func (lpp *MSubmoduleLPPProver) proveLPP(req MLPPRequest) (MLPPResponse, error) {
	dummyproof := MLPPResponse{}
	return dummyproof, nil
}

// MExecConglomerator combines various proofs into a final execution proof
type MExecConglomerator struct {
}

// prove combines GL and LPP responses into a final execution proof
func (cong *MExecConglomerator) prove(glresp MGLResp, lppresp MLPPResponse) (execution.Response, error) {
	return execution.Response{}, nil
}
