package coinbasemanager

import (
	"strconv"
	"testing"

	"github.com/Nautilus-Network/nautiliad/domain/consensus/model/externalapi"
	"github.com/Nautilus-Network/nautiliad/domain/consensus/utils/constants"
	"github.com/Nautilus-Network/nautiliad/domain/dagconfig"
)

func TestCalcHalvingPeriodBlockSubsidy(t *testing.T) {
	const secondsPerMonth = 2629800
	const secondsPerHalving = secondsPerMonth * 12
	const halvingPhaseDaaScore = secondsPerMonth * 12
	const halvingPhaseBaseSubsidy = 6 * constants.SompiPerKaspa
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		halvingPhaseDaaScore,
		halvingPhaseBaseSubsidy,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
	coinbaseManagerInstance := coinbaseManagerInterface.(*coinbaseManager)

	tests := []struct {
		name                 string
		blockDaaScore        uint64
		expectedBlockSubsidy uint64
	}{
		{
			name:                 "start of halving phase",
			blockDaaScore:        halvingPhaseDaaScore,
			expectedBlockSubsidy: halvingPhaseBaseSubsidy,
		},
		{
			name:                 "after one halving",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving,
			expectedBlockSubsidy: halvingPhaseBaseSubsidy / 2,
		},
		{
			name:                 "after two halvings",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving*2,
			expectedBlockSubsidy: halvingPhaseBaseSubsidy / 4,
		},
		{
			name:                 "after five halvings",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving*5,
			expectedBlockSubsidy: halvingPhaseBaseSubsidy / 32,
		},
		{
			name:                 "after 32 halvings",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving*32,
			expectedBlockSubsidy: halvingPhaseBaseSubsidy / 4294967296,
		},
		{
			name:                 "just before subsidy depleted",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving*35,
			expectedBlockSubsidy: 1,
		},
		{
			name:                 "after subsidy depleted",
			blockDaaScore:        halvingPhaseDaaScore + secondsPerHalving*36,
			expectedBlockSubsidy: 0,
		},
	}

	for _, test := range tests {
		blockSubsidy := coinbaseManagerInstance.calcHalvingPeriodBlockSubsidy(test.blockDaaScore)
		if blockSubsidy != test.expectedBlockSubsidy {
			t.Errorf("TestCalcHalvingPeriodBlockSubsidy: test '%s' failed. Want: %d, got: %d",
				test.name, test.expectedBlockSubsidy, blockSubsidy)
		}
	}
}

func TestBuildSubsidyTable(t *testing.T) {
	halvingPhaseBaseSubsidy := dagconfig.MainnetParams.HalvingPhaseBaseSubsidy
	if halvingPhaseBaseSubsidy != 8*constants.SompiPerKaspa {
		t.Errorf("TestBuildSubsidyTable: table generation function was not updated to reflect "+
			"the new base subsidy %d. Please fix the constant above and replace subsidyByHalvingMonthTable "+
			"in coinbasemanager.go with the printed table", halvingPhaseBaseSubsidy)
	}
	coinbaseManagerInterface := New(
		nil,
		0,
		0,
		0,
		&externalapi.DomainHash{},
		0,
		halvingPhaseBaseSubsidy,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil)
	coinbaseManagerInstance := coinbaseManagerInterface.(*coinbaseManager)

	var subsidyTable []uint64
	for M := uint64(0); ; M++ {
		subsidy := coinbaseManagerInstance.calcHalvingPeriodBlockSubsidyFloatCalc(M)
		subsidyTable = append(subsidyTable, subsidy)
		if subsidy == 0 {
			break
		}
	}

	tableStr := "\n{\t"
	for i := 0; i < len(subsidyTable); i++ {
		tableStr += strconv.FormatUint(subsidyTable[i], 10) + ", "
		if (i+1)%25 == 0 {
			tableStr += "\n\t"
		}
	}
	tableStr += "\n}"
	t.Logf(tableStr)
}
