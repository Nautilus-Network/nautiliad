package coinbasemanager

import (
	"math"

	"github.com/Nautilus-Network/nautiliad/domain/consensus/model"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/model/externalapi"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/utils/constants"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/utils/hashset"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/utils/subnetworks"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/utils/transactionhelper"
	"github.com/Nautilus-Network/nautiliad/infrastructure/db/database"
	"github.com/pkg/errors"
)

type coinbaseManager struct {
	subsidyGenesisReward                    uint64
	preHalvingPhaseBaseSubsidy              uint64
	coinbasePayloadScriptPublicKeyMaxLength uint8
	genesisHash                             *externalapi.DomainHash
	halvingPhaseDaaScore                    uint64
	halvingPhaseBaseSubsidy                 uint64

	databaseContext     model.DBReader
	dagTraversalManager model.DAGTraversalManager
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
	daaBlocksStore      model.DAABlocksStore
	blockStore          model.BlockStore
	pruningStore        model.PruningStore
	blockHeaderStore    model.BlockHeaderStore
}

func (c *coinbaseManager) ExpectedCoinbaseTransaction(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData) (expectedTransaction *externalapi.DomainTransaction, hasRedReward bool, err error) {

	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, true)
	if !database.IsNotFoundError(err) && err != nil {
		return nil, false, err
	}

	// If there's ghostdag data with trusted data we prefer it because we need the original merge set non-pruned merge set.
	if database.IsNotFoundError(err) {
		ghostdagData, err = c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
		if err != nil {
			return nil, false, err
		}
	}

	acceptanceData, err := c.acceptanceDataStore.Get(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	daaAddedBlocksSet, err := c.daaAddedBlocksSet(stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	txOuts := make([]*externalapi.DomainTransactionOutput, 0, len(ghostdagData.MergeSetBlues()))
	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	for _, blue := range ghostdagData.MergeSetBlues() {
		txOut, hasReward, err := c.coinbaseOutputForBlueBlock(stagingArea, blue, acceptanceDataMap[*blue], daaAddedBlocksSet)
		if err != nil {
			return nil, false, err
		}

		if hasReward {
			txOuts = append(txOuts, txOut)
		}
	}

	txOut, hasRedReward, err := c.coinbaseOutputForRewardFromRedBlocks(
		stagingArea, ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
	if err != nil {
		return nil, false, err
	}

	if hasRedReward {
		txOuts = append(txOuts, txOut)
	}

	subsidy, err := c.CalcBlockSubsidy(stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	payload, err := c.serializeCoinbasePayload(ghostdagData.BlueScore(), coinbaseData, subsidy)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      txOuts,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		Payload:      payload,
	}, hasRedReward, nil
}

func (c *coinbaseManager) daaAddedBlocksSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	hashset.HashSet, error) {

	daaAddedBlocks, err := c.daaBlocksStore.DAAAddedBlocks(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return hashset.NewFromSlice(daaAddedBlocks...), nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputForBlueBlock(stagingArea *model.StagingArea,
	blueBlock *externalapi.DomainHash, blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, bool, error) {

	blockReward, err := c.calcMergedBlockReward(stagingArea, blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
	if err != nil {
		return nil, false, err
	}

	if blockReward == 0 {
		return nil, false, nil
	}

	// the ScriptPublicKey for the coinbase is parsed from the coinbase payload
	_, coinbaseData, _, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(blockAcceptanceData.TransactionAcceptanceData[0].Transaction)
	if err != nil {
		return nil, false, err
	}

	txOut := &externalapi.DomainTransactionOutput{
		Value:           blockReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}

	return txOut, true, nil
}

func (c *coinbaseManager) coinbaseOutputForRewardFromRedBlocks(stagingArea *model.StagingArea,
	ghostdagData *externalapi.BlockGHOSTDAGData, acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, bool, error) {

	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	totalReward := uint64(0)
	for _, red := range ghostdagData.MergeSetReds() {
		reward, err := c.calcMergedBlockReward(stagingArea, red, acceptanceDataMap[*red], daaAddedBlocksSet)
		if err != nil {
			return nil, false, err
		}

		totalReward += reward
	}

	if totalReward == 0 {
		return nil, false, nil
	}

	return &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}, true, nil
}

func acceptanceDataFromArrayToMap(acceptanceData externalapi.AcceptanceData) map[externalapi.DomainHash]*externalapi.BlockAcceptanceData {
	acceptanceDataMap := make(map[externalapi.DomainHash]*externalapi.BlockAcceptanceData, len(acceptanceData))
	for _, blockAcceptanceData := range acceptanceData {
		acceptanceDataMap[*blockAcceptanceData.BlockHash] = blockAcceptanceData
	}
	return acceptanceDataMap
}

// CalcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
func (c *coinbaseManager) CalcBlockSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	if blockHash.Equal(c.genesisHash) {
		return c.subsidyGenesisReward, nil
	}
	blockDaaScore, err := c.daaBlocksStore.DAAScore(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	if blockDaaScore < c.halvingPhaseDaaScore {
		return c.preHalvingPhaseBaseSubsidy, nil
	}

	blockSubsidy := c.calcHalvingPeriodBlockSubsidy(blockDaaScore)
	return blockSubsidy, nil
}

func (c *coinbaseManager) calcHalvingPeriodBlockSubsidy(blockDaaScore uint64) uint64 {
	// We define a year as 365.25 days and a month as 365.25 / 12 = 30.4375
	// secondsPerMonth = 30.4375 * 24 * 60 * 60 = 2629800
	// secondsPerHalving = 2629800 * 12
	const secondsPerHalving = 31557600
	// Note that this calculation implicitly assumes that block per second = 1 (by assuming daa score diff is in second units).
	monthsSinceHalvingPhaseStarted := (blockDaaScore - c.halvingPhaseDaaScore) / secondsPerHalving
	// monthsSinceHalvingHalvingPhaseStarted := (blockDaaScore - c.halvingPhaseDaaScore) / secondsPerMonth
	// Return the pre-calculated value from subsidy-per-month table
	return c.getHalvingPeriodBlockSubsidyFromTable(monthsSinceHalvingPhaseStarted)
}

/*
This table was pre-calculated by calling `calcHalvingPeriodBlockSubsidyFloatCalc` for all months until reaching 0 subsidy.
To regenerate this table, run `TestBuildSubsidyTable` in coinbasemanager_test.go (note the `halvingPhaseBaseSubsidy` therein)
*/
var subsidyByHalvingMonthTable = []uint64{
	600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 600000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 300000000, 150000000,
	150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 150000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 75000000, 37500000, 37500000,
	37500000, 37500000, 37500000, 37500000, 37500000, 37500000, 37500000, 37500000, 37500000, 37500000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 18750000, 9375000, 9375000, 9375000,
	9375000, 9375000, 9375000, 9375000, 9375000, 9375000, 9375000, 9375000, 9375000, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 4687500, 2343750, 2343750, 2343750, 2343750,
	2343750, 2343750, 2343750, 2343750, 2343750, 2343750, 2343750, 2343750, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 1171875, 5859375, 5859375, 5859375, 5859375, 5859375,
	5859375, 5859375, 5859375, 5859375, 5859375, 5859375, 5859375, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 2929688, 1464844, 1464844, 1464844, 1464844, 1464844, 1464844,
	1464844, 1464844, 1464844, 1464844, 1464844, 1464844, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 732422, 366211, 366211, 366211, 366211, 366211, 366211, 366211,
	366211, 366211, 366211, 366211, 366211, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 183105, 91553, 91553, 91553, 91553, 91553, 91553, 91553, 91553,
	91553, 91553, 91553, 91553, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 45776, 22888, 22888, 22888, 22888, 22888, 22888, 22888, 22888, 22888,
	22888, 22888, 22888, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 11444, 5722, 5722, 5722, 5722, 5722, 5722, 5722, 5722, 5722, 5722,
	5722, 5722, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 2861, 1431, 1431, 1431, 1431, 1431, 1431, 1431, 1431, 1431, 1431, 1431,
	1431, 715, 715, 715, 715, 715, 715, 715, 715, 715, 715, 715, 715, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358,
	179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 45,
	45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	0,
}

func (c *coinbaseManager) getHalvingPeriodBlockSubsidyFromTable(month uint64) uint64 {
	if month >= uint64(len(subsidyByHalvingMonthTable)) {
		month = uint64(len(subsidyByHalvingMonthTable) - 1)
	}
	return subsidyByHalvingMonthTable[month]
}

func (c *coinbaseManager) calcHalvingPeriodBlockSubsidyFloatCalc(month uint64) uint64 {
	baseSubsidy := c.halvingPhaseBaseSubsidy
	subsidy := float64(baseSubsidy) / math.Pow(2, float64(month)/12)
	return uint64(subsidy)
}

func (c *coinbaseManager) calcMergedBlockReward(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData, mergingBlockDAAAddedBlocksSet hashset.HashSet) (uint64, error) {

	if !blockHash.Equal(blockAcceptanceData.BlockHash) {
		return 0, errors.Errorf("blockAcceptanceData.BlockHash is expected to be %s but got %s",
			blockHash, blockAcceptanceData.BlockHash)
	}

	if !mergingBlockDAAAddedBlocksSet.Contains(blockHash) {
		return 0, nil
	}

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	block, err := c.blockStore.Block(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, err
	}

	_, _, subsidy, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(block.Transactions[transactionhelper.CoinbaseTransactionIndex])
	if err != nil {
		return 0, err
	}

	return subsidy + totalFees, nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyGenesisReward uint64,
	preHalvingPhaseBaseSubsidy uint64,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	genesisHash *externalapi.DomainHash,
	halvingPhaseDaaScore uint64,
	halvingPhaseBaseSubsidy uint64,

	dagTraversalManager model.DAGTraversalManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore,
	daaBlocksStore model.DAABlocksStore,
	blockStore model.BlockStore,
	pruningStore model.PruningStore,
	blockHeaderStore model.BlockHeaderStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyGenesisReward:                    subsidyGenesisReward,
		preHalvingPhaseBaseSubsidy:              preHalvingPhaseBaseSubsidy,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		genesisHash:                             genesisHash,
		halvingPhaseDaaScore:                    halvingPhaseDaaScore,
		halvingPhaseBaseSubsidy:                 halvingPhaseBaseSubsidy,

		dagTraversalManager: dagTraversalManager,
		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
		daaBlocksStore:      daaBlocksStore,
		blockStore:          blockStore,
		pruningStore:        pruningStore,
		blockHeaderStore:    blockHeaderStore,
	}
}
