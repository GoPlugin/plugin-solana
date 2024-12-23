package smoke

import (
	"fmt"
	"maps"
	"os/exec"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/goplugin/pluginv3.0/integration-tests/docker/test_env"

	"github.com/goplugin/plugin-solana/integration-tests/common"
	ocr_config "github.com/goplugin/plugin-solana/integration-tests/config"
	"github.com/goplugin/plugin-solana/integration-tests/gauntlet"
	tc "github.com/goplugin/plugin-solana/integration-tests/testconfig"
	"github.com/goplugin/plugin-solana/integration-tests/utils"
)

func TestSolanaOCRV2Smoke(t *testing.T) {
	for _, test := range []struct {
		name string
		env  map[string]string
	}{
		{name: "embedded"},
		{name: "plugins", env: map[string]string{
			"CL_MEDIAN_CMD": "plugin-feeds",
			"CL_SOLANA_CMD": "plugin-solana",
		}},
	} {
		config, err := tc.GetConfig("Smoke", tc.OCR2)
		if err != nil {
			t.Fatal(err)
		}

		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, sg := startOCR2DataFeedsSmokeTest(t, test.name, test.env, config, "")
			validateRounds(t, test.name, sg, *config.OCR2.NumberOfRounds)
		})
	}
}

func startOCR2DataFeedsSmokeTest(t *testing.T, testname string, testenv map[string]string, config tc.TestConfig, subDir string) (*common.OCRv2TestState, *gauntlet.SolanaGauntlet) {
	name := "gauntlet-" + testname
	state, err := common.NewOCRv2State(t, 1, name, &config)
	require.NoError(t, err, "Could not setup the ocrv2 state")
	if len(testenv) > 0 {
		state.Common.TestEnvDetails.NodeOpts = append(state.Common.TestEnvDetails.NodeOpts, func(n *test_env.ClNode) {
			if n.ContainerEnvs == nil {
				n.ContainerEnvs = map[string]string{}
			}
			maps.Copy(n.ContainerEnvs, testenv)
		})
	}

	state.DeployCluster(utils.ContractsDir)
	state.DeployContracts(utils.ContractsDir, subDir)
	if state.Common.Env.WillUseRemoteRunner() {
		return state, nil
	}

	// copy gauntlet folder to run in parallel (gauntlet generates an output file that is read by the e2e tests - causes conflict if shared)
	gauntletCopyPath := utils.ProjectRoot + "/" + name
	if out, cpErr := exec.Command("cp", "-r", utils.ProjectRoot+"/gauntlet", gauntletCopyPath).Output(); cpErr != nil { // nolint:gosec
		require.NoError(t, err, "output: "+string(out))
	}

	sg, err := gauntlet.NewSolanaGauntlet(gauntletCopyPath)
	require.NoError(t, err)
	state.Gauntlet = sg

	if *config.Common.InsideK8s {
		t.Cleanup(func() {
			err = state.Common.Env.Shutdown()
			if err != nil {
				log.Err(err)
			}
		})
	}

	state.SetupClients()
	require.NoError(t, err)

	gauntletConfig := map[string]string{
		"SECRET":      fmt.Sprintf("\"%s\"", *config.SolanaConfig.Secret),
		"NODE_URL":    state.Common.ChainDetails.RPCURLExternal,
		"WS_URL":      state.Common.ChainDetails.WSURLExternal,
		"PRIVATE_KEY": state.Common.AccountDetails.PrivateKey,
	}

	err = sg.SetupNetwork(gauntletConfig)
	require.NoError(t, err, "Error setting gauntlet network")
	err = sg.InstallDependencies()
	require.NoError(t, err, "Error installing gauntlet dependencies")

	if *config.Common.Network == "devnet" {
		state.Common.ChainDetails.ProgramAddresses.OCR2 = *config.SolanaConfig.OCR2ProgramID
		state.Common.ChainDetails.ProgramAddresses.AccessController = *config.SolanaConfig.AccessControllerProgramID
		state.Common.ChainDetails.ProgramAddresses.Store = *config.SolanaConfig.StoreProgramID
		sg.LinkAddress = *config.SolanaConfig.LinkTokenAddress
		sg.VaultAddress = *config.SolanaConfig.VaultAddress
	} else {
		// Deploying PLI in case of localnet
		err = sg.DeployLinkToken()
		require.NoError(t, err)
	}

	err = sg.G.WriteNetworkConfigVar(sg.NetworkFilePath, "PROGRAM_ID_OCR2", state.Common.ChainDetails.ProgramAddresses.OCR2)
	require.NoError(t, err, "Error adding gauntlet variable")
	err = sg.G.WriteNetworkConfigVar(sg.NetworkFilePath, "PROGRAM_ID_ACCESS_CONTROLLER", state.Common.ChainDetails.ProgramAddresses.AccessController)
	require.NoError(t, err, "Error adding gauntlet variable")
	err = sg.G.WriteNetworkConfigVar(sg.NetworkFilePath, "PROGRAM_ID_STORE", state.Common.ChainDetails.ProgramAddresses.Store)
	require.NoError(t, err, "Error adding gauntlet variable")
	err = sg.G.WriteNetworkConfigVar(sg.NetworkFilePath, "PLI", sg.LinkAddress)
	require.NoError(t, err, "Error adding gauntlet variable")
	err = sg.G.WriteNetworkConfigVar(sg.NetworkFilePath, "VAULT_ADDRESS", sg.VaultAddress)
	require.NoError(t, err, "Error adding gauntlet variable")

	_, err = sg.DeployOCR2()
	require.NoError(t, err, "Error deploying OCR")
	// Generating default OCR2 config
	ocr2Config := ocr_config.NewOCR2Config(state.Clients.PluginClient.NKeys, sg.ProposalAddress, sg.VaultAddress, *config.SolanaConfig.Secret)
	ocr2Config.Default()
	sg.OCR2Config = ocr2Config

	err = sg.ConfigureOCR2()
	require.NoError(t, err)

	state.CreateJobs()
	return state, sg
}

func validateRounds(t *testing.T, testname string, sg *gauntlet.SolanaGauntlet, rounds int) {
	name := "gauntlet" + testname

	// Test start
	stuck := 0
	successFullRounds := 0
	prevRound := gauntlet.Transmission{
		RoundID: 0,
	}
	for successFullRounds < rounds {
		time.Sleep(time.Second * 6)
		require.Less(t, stuck, 10, fmt.Sprintf("%s: Rounds have been stuck for more than 10 iterations", name))
		log.Info().Str("Transmission", sg.OcrAddress).Msg("Inspecting transmissions")
		transmissions, err := sg.FetchTransmissions(sg.OcrAddress)
		require.NoError(t, err)
		if len(transmissions) <= 1 {
			log.Info().Str("Contract", sg.OcrAddress).Msg(fmt.Sprintf("%s: No Transmissions", name))
			stuck++
			continue
		}
		currentRound := common.GetLatestRound(transmissions)
		if prevRound.RoundID == 0 {
			prevRound = currentRound
		}
		if currentRound.RoundID <= prevRound.RoundID {
			log.Info().Str("Transmission", sg.OcrAddress).Msg(fmt.Sprintf("%s: No new transmissions", name))
			stuck++
			continue
		}
		log.Info().Str("Contract", sg.OcrAddress).Interface("Answer", currentRound.Answer).Int64("RoundID", currentRound.RoundID).Msg(fmt.Sprintf("%s: New answer found", name))
		require.Equal(t, currentRound.Answer, int64(5), fmt.Sprintf("Actual: %d, Expected: 5", currentRound.Answer))
		require.Less(t, prevRound.RoundID, currentRound.RoundID, fmt.Sprintf("Expected round %d to be less than %d", prevRound.RoundID, currentRound.RoundID))
		prevRound = currentRound
		successFullRounds++
		stuck = 0
	}
}
