package common

import (
	"fmt"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	test_env_ctf "github.com/goplugin/plugin-testing-framework/lib/docker/test_env"
	"github.com/goplugin/plugin-testing-framework/lib/utils/testcontext"

	"github.com/goplugin/pluginv3.0/integration-tests/client"
	"github.com/goplugin/pluginv3.0/integration-tests/docker/test_env"

	"github.com/goplugin/pluginv3.0/v2/core/services/job"
	"github.com/goplugin/pluginv3.0/v2/core/store/models"

	test_env_sol "github.com/goplugin/plugin-solana/integration-tests/docker/testenv"
	"github.com/goplugin/plugin-solana/integration-tests/gauntlet"
	"github.com/goplugin/plugin-solana/integration-tests/solclient"
	"github.com/goplugin/plugin-solana/integration-tests/testconfig"
)

type OCRv2TestState struct {
	ContractDeployer   *solclient.ContractDeployer
	LinkToken          *solclient.LinkToken
	ContractsNodeSetup map[int]*ContractNodeInfo
	Clients            *Clients
	Common             *Common
	Config             *Config
	Gauntlet           *gauntlet.SolanaGauntlet
}

type Clients struct {
	SolanaClient    *solclient.Client
	KillgraveClient *test_env_ctf.Killgrave
	PluginClient *PluginClient
}

type PluginClient struct {
	PluginClientDocker *test_env.ClCluster
	PluginClientK8s    []*client.PluginK8sClient
	PluginNodes        []*client.PluginClient
	NKeys                 []client.NodeKeysBundle
	AccountAddresses      []string
}

type Config struct {
	T          *testing.T
	TestConfig *testconfig.TestConfig
	Resty      *resty.Client
	err        error
}

func NewOCRv2State(t *testing.T, contracts int, namespacePrefix string, testConfig *testconfig.TestConfig) (*OCRv2TestState, error) {
	c, err := New(testConfig).Default(t, namespacePrefix)
	if err != nil {
		return nil, err
	}
	state := &OCRv2TestState{
		ContractsNodeSetup: make(map[int]*ContractNodeInfo),
		Common:             c,
		Clients: &Clients{
			SolanaClient:    &solclient.Client{},
			PluginClient: &PluginClient{},
		},
		Config: &Config{
			T:          t,
			TestConfig: testConfig,
			Resty:      nil,
			err:        nil,
		},
	}

	state.Clients.SolanaClient.Config = state.Clients.SolanaClient.Config.Default()
	for i := 0; i < contracts; i++ {
		state.ContractsNodeSetup[i] = &ContractNodeInfo{}
		state.ContractsNodeSetup[i].BootstrapNodeIdx = 0
		for n := 1; n < *state.Config.TestConfig.OCR2.NodeCount; n++ {
			state.ContractsNodeSetup[i].NodesIdx = append(state.ContractsNodeSetup[i].NodesIdx, n)
		}
	}
	return state, nil
}

type ContractsState struct {
	OCR           string `json:"ocr"`
	Store         string `json:"store"`
	Feed          string `json:"feed"`
	Owner         string `json:"owner"`
	Mint          string `json:"mint"`
	MintAuthority string `json:"mint_authority"`
	OCRVault      string `json:"ocr_vault"`
}

func (m *OCRv2TestState) DeployCluster(contractsDir string) {
	if *m.Config.TestConfig.Common.InsideK8s {
		m.DeployEnv(contractsDir)

		if m.Common.Env.WillUseRemoteRunner() {
			return
		}

		// Setting up the URLs
		m.Common.ChainDetails.RPCURLExternal = m.Common.Env.URLs["sol"][0]
		m.Common.ChainDetails.WSURLExternal = m.Common.Env.URLs["sol"][1]

		if *m.Config.TestConfig.Common.Network == "devnet" {
			m.Common.ChainDetails.RPCUrl = *m.Config.TestConfig.Common.RPCURL
			m.Common.ChainDetails.RPCURLExternal = *m.Config.TestConfig.Common.RPCURL
			m.Common.ChainDetails.WSURLExternal = *m.Config.TestConfig.Common.WsURL
		}

		m.Common.ChainDetails.MockserverURLInternal = m.Common.Env.URLs["qa_mock_adapter_internal"][0]
		m.Common.ChainDetails.MockServerEndpoint = "five"
	} else {
		env, err := test_env.NewTestEnv()
		require.NoError(m.Config.T, err)
		sol := test_env_sol.NewSolana([]string{env.DockerNetwork.Name}, *m.Config.TestConfig.Common.DevnetImage, m.Common.AccountDetails.PublicKey)
		err = sol.StartContainer()
		require.NoError(m.Config.T, err)

		// Setting the External RPC url for Gauntlet
		m.Common.ChainDetails.RPCUrl = sol.InternalHTTPURL
		m.Common.ChainDetails.RPCURLExternal = sol.ExternalHTTPURL
		m.Common.ChainDetails.WSURLExternal = sol.ExternalWsURL

		if *m.Config.TestConfig.Common.Network == "devnet" {
			m.Common.ChainDetails.RPCUrl = *m.Config.TestConfig.Common.RPCURL
			m.Common.ChainDetails.RPCURLExternal = *m.Config.TestConfig.Common.RPCURL
			m.Common.ChainDetails.WSURLExternal = *m.Config.TestConfig.Common.WsURL
		}

		b, err := test_env.NewCLTestEnvBuilder().
			WithNonEVM().
			WithTestInstance(m.Config.T).
			WithTestConfig(m.Config.TestConfig).
			WithMockAdapter().
			WithCLNodes(*m.Config.TestConfig.OCR2.NodeCount).
			WithCLNodeOptions(m.Common.TestEnvDetails.NodeOpts...).
			WithStandardCleanup().
			WithTestEnv(env)
		require.NoError(m.Config.T, err)
		env, err = b.Build()
		require.NoError(m.Config.T, err)
		m.Common.DockerEnv = &SolCLClusterTestEnv{
			CLClusterTestEnv: env,
			Sol:              sol,
			Killgrave:        env.MockAdapter,
		}
		// Setting up Mock adapter
		m.Clients.KillgraveClient = env.MockAdapter
		m.Common.ChainDetails.MockserverURLInternal = m.Clients.KillgraveClient.InternalEndpoint
		m.Common.ChainDetails.MockServerEndpoint = "mockserver-bridge"
		err = m.Clients.KillgraveClient.SetAdapterBasedIntValuePath("/mockserver-bridge", []string{http.MethodGet, http.MethodPost}, 5)
		require.NoError(m.Config.T, err, "Failed to set mock adapter value")
	}

	m.SetupClients()
	m.SetPluginNodes()
}

// UploadProgramBinaries uploads programs binary files to solana-validator container
// currently it's the only way to deploy anything to local solana because ephemeral validator in k8s
// can't expose UDP ports required to copy .so chunks when deploying
func (m *OCRv2TestState) UploadProgramBinaries(contractsDir string) {
	pl, err := m.Common.Env.Client.ListPods(m.Common.Env.Cfg.Namespace, "app=sol")
	require.NoError(m.Config.T, err)
	_, _, _, err = m.Common.Env.Client.CopyToPod(m.Common.Env.Cfg.Namespace, contractsDir, fmt.Sprintf("%s/%s:/programs", m.Common.Env.Cfg.Namespace, pl.Items[0].Name), "sol-val")
	require.NoError(m.Config.T, err)
}

func (m *OCRv2TestState) DeployEnv(contractsDir string) {
	err := m.Common.Env.Run()
	require.NoError(m.Config.T, err)

	if !m.Common.Env.WillUseRemoteRunner() {
		m.UploadProgramBinaries(contractsDir)
	}
}

func (m *OCRv2TestState) NewSolanaClientSetup(networkSettings *solclient.SolNetwork) (*solclient.Client, error) {
	if *m.Config.TestConfig.Common.InsideK8s {
		networkSettings.URLs = m.Common.Env.URLs[networkSettings.Name]
	} else {
		networkSettings.URLs = []string{
			m.Common.DockerEnv.Sol.ExternalHTTPURL,
			m.Common.DockerEnv.Sol.ExternalWsURL,
		}
	}
	ec, err := solclient.NewClient(networkSettings)
	if err != nil {
		return nil, err
	}
	log.Info().
		Interface("URLs", networkSettings.URLs).
		Msg("Connected Solana client")
	return ec, nil
}

func (m *OCRv2TestState) SetupClients() {
	solClient, err := m.NewSolanaClientSetup(m.Clients.SolanaClient.Config)
	m.Clients.SolanaClient = solClient
	require.NoError(m.Config.T, err)
	if *m.Config.TestConfig.Common.InsideK8s {
		m.Clients.PluginClient.PluginClientK8s, err = client.ConnectPluginNodes(m.Common.Env)
		require.NoError(m.Config.T, err)
	} else {
		m.Clients.PluginClient.PluginClientDocker = m.Common.DockerEnv.CLClusterTestEnv.ClCluster
	}
}

// DeployContracts deploys contracts
// baseDir is the root folder where contracts are stored
// subDir allows for pointing to a subdirectory within baseDir (can be left empty)
func (m *OCRv2TestState) DeployContracts(baseDir, subDir string) {
	var err error
	m.Clients.PluginClient.NKeys, err = m.Common.CreateNodeKeysBundle(m.Clients.PluginClient.PluginNodes)
	require.NoError(m.Config.T, err)
	cd, err := solclient.NewContractDeployer(m.Clients.SolanaClient, nil)
	require.NoError(m.Config.T, err)
	if *m.Config.TestConfig.Common.InsideK8s {
		err = cd.DeployAnchorProgramsRemote(baseDir, m.Common.Env)
	} else {
		err = cd.DeployAnchorProgramsRemoteDocker(baseDir, subDir, m.Common.DockerEnv.Sol, solclient.BuildProgramIDKeypairPath)
	}
	require.NoError(m.Config.T, err)
}

func (m *OCRv2TestState) UpgradeContracts(baseDir, subDir string) {
	cd, err := solclient.NewContractDeployer(m.Clients.SolanaClient, nil)
	require.NoError(m.Config.T, err)

	// fetch corresponding program address for program
	programIDBuilder := func(programName string) string {
		// remove extra directories + .so suffix from lookup
		programName, _ = strings.CutSuffix(filepath.Base(programName), ".so")
		ids := map[string]string{
			"ocr_2":             m.Common.ChainDetails.ProgramAddresses.OCR2,
			"access_controller": m.Common.ChainDetails.ProgramAddresses.AccessController,
			"store":             m.Common.ChainDetails.ProgramAddresses.Store,
		}
		val, ok := ids[programName]
		require.True(m.Config.T, ok, fmt.Sprintf("unable to find corresponding key (%s) within %+v", programName, ids))
		return val
	}

	if *m.Config.TestConfig.Common.InsideK8s {
		err = fmt.Errorf("not implemented")
	} else {
		err = cd.DeployAnchorProgramsRemoteDocker(baseDir, subDir, m.Common.DockerEnv.Sol, programIDBuilder)
	}
	require.NoError(m.Config.T, err)
}

// CreateJobs creating OCR jobs and EA stubs
func (m *OCRv2TestState) CreateJobs() {
	// Setting up RPC used for external network funding
	c := rpc.New(m.Common.ChainDetails.RPCURLExternal)
	wsc, err := ws.Connect(testcontext.Get(m.Config.T), m.Common.ChainDetails.WSURLExternal)
	require.NoError(m.Config.T, err, "Error connecting to websocket client")

	relayConfig := job.JSONConfig{
		"nodeEndpointHTTP": m.Common.ChainDetails.RPCUrl,
		"ocr2ProgramID":    m.Common.ChainDetails.ProgramAddresses.OCR2,
		"transmissionsID":  m.Gauntlet.FeedAddress,
		"storeProgramID":   m.Common.ChainDetails.ProgramAddresses.Store,
		"chainID":          m.Common.ChainDetails.ChainID,
	}
	boostratInternalIP := m.Clients.PluginClient.PluginNodes[0].InternalIP()
	bootstrapPeers := []client.P2PData{
		{
			InternalIP:   boostratInternalIP,
			InternalPort: "6690",
			PeerID:       m.Clients.PluginClient.NKeys[0].PeerID,
		},
	}
	jobSpec := &client.OCR2TaskJobSpec{
		Name:    fmt.Sprintf("sol-OCRv2-%s-%s", "bootstrap", uuid.New().String()),
		JobType: "bootstrap",
		OCR2OracleSpec: job.OCR2OracleSpec{
			ContractID:                        m.Gauntlet.OcrAddress,
			Relay:                             m.Common.ChainDetails.ChainName,
			RelayConfig:                       relayConfig,
			P2PV2Bootstrappers:                pq.StringArray{bootstrapPeers[0].P2PV2Bootstrapper()},
			OCRKeyBundleID:                    null.StringFrom(m.Clients.PluginClient.NKeys[0].OCR2Key.Data.ID),
			TransmitterID:                     null.StringFrom(m.Clients.PluginClient.NKeys[0].TXKey.Data.ID),
			ContractConfigConfirmations:       1,
			ContractConfigTrackerPollInterval: models.Interval(15 * time.Second),
		},
	}
	sourceValueBridge := client.BridgeTypeAttributes{
		Name:        "mockserver-bridge",
		URL:         fmt.Sprintf("%s/%s", m.Common.ChainDetails.MockserverURLInternal, m.Common.ChainDetails.MockServerEndpoint),
		RequestData: "{}",
	}

	observationSource := client.ObservationSourceSpecBridge(&sourceValueBridge)
	bridgeInfo := BridgeInfo{ObservationSource: observationSource}

	err = m.Clients.PluginClient.PluginNodes[0].MustCreateBridge(&sourceValueBridge)
	require.NoError(m.Config.T, err, "Error creating bridge")

	_, err = m.Clients.PluginClient.PluginNodes[0].MustCreateJob(jobSpec)
	require.NoError(m.Config.T, err, "Error creating job")

	for nIdx, node := range m.Clients.PluginClient.PluginNodes {
		// Skipping bootstrap
		if nIdx == 0 {
			continue
		}
		if *m.Config.TestConfig.Common.Network == "localnet" {
			err = m.Clients.SolanaClient.Fund(m.Clients.PluginClient.NKeys[nIdx].TXKey.Data.ID, big.NewFloat(1e4))
			require.NoError(m.Config.T, err, "Error sending funds")
		} else {
			err = solclient.SendFunds(*m.Config.TestConfig.Common.PrivateKey, m.Clients.PluginClient.NKeys[nIdx].TXKey.Data.ID, 100000000, c, wsc)
			require.NoError(m.Config.T, err, "Error sending funds")
		}

		sourceValueBridge := client.BridgeTypeAttributes{
			Name:        "mockserver-bridge",
			URL:         fmt.Sprintf("%s/%s", m.Common.ChainDetails.MockserverURLInternal, m.Common.ChainDetails.MockServerEndpoint),
			RequestData: "{}",
		}

		_, err := node.CreateBridge(&sourceValueBridge)
		require.NoError(m.Config.T, err, "Error creating bridge")

		jobSpec := &client.OCR2TaskJobSpec{
			Name:              fmt.Sprintf("sol-OCRv2-%d-%s", nIdx, uuid.New().String()),
			JobType:           "offchainreporting2",
			ObservationSource: bridgeInfo.ObservationSource,
			OCR2OracleSpec: job.OCR2OracleSpec{
				ContractID:                        m.Gauntlet.OcrAddress,
				Relay:                             m.Common.ChainDetails.ChainName,
				RelayConfig:                       relayConfig,
				P2PV2Bootstrappers:                pq.StringArray{bootstrapPeers[0].P2PV2Bootstrapper()},
				OCRKeyBundleID:                    null.StringFrom(m.Clients.PluginClient.NKeys[nIdx].OCR2Key.Data.ID),
				TransmitterID:                     null.StringFrom(m.Clients.PluginClient.NKeys[nIdx].TXKey.Data.ID),
				ContractConfigConfirmations:       1,
				ContractConfigTrackerPollInterval: models.Interval(15 * time.Second),
				PluginType:                        "median",
				PluginConfig:                      PluginConfigToTomlFormat(observationSource),
			},
		}
		_, err = node.MustCreateJob(jobSpec)
		require.NoError(m.Config.T, err, "Error creating job")
	}
}

func (m *OCRv2TestState) SetPluginNodes() {
	// retrieve client from K8s client
	pluginNodes := []*client.PluginClient{}
	if *m.Config.TestConfig.Common.InsideK8s {
		for i := range m.Clients.PluginClient.PluginClientK8s {
			pluginNodes = append(pluginNodes, m.Clients.PluginClient.PluginClientK8s[i].PluginClient)
		}
	} else {
		pluginNodes = append(pluginNodes, m.Clients.PluginClient.PluginClientDocker.NodeAPIs()...)
	}
	m.Clients.PluginClient.PluginNodes = pluginNodes
}

func formatBuffer(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}
	result := fmt.Sprintf("%d", buf[0])
	for _, b := range buf[1:] {
		result += fmt.Sprintf(",%d", b)
	}
	return result
}

func GetLatestRound(transmissions []gauntlet.Transmission) gauntlet.Transmission {
	highestRound := transmissions[0]
	for _, t := range transmissions[1:] {
		if t.RoundID > highestRound.RoundID {
			highestRound = t
		}
	}
	return highestRound
}
