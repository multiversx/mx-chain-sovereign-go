package rating

import (
	"github.com/multiversx/mx-chain-go/statusHandler"
	"golang.org/x/exp/slices"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignRatingsData struct {
	*RatingsData
}

// NewSovereignRatingsData creates a sovereign ratings data
func NewSovereignRatingsData(args RatingsDataArg) (*sovereignRatingsData, error) {
	ratingsConfig := args.Config
	err := verifySovereignRatingsConfig(ratingsConfig)
	if err != nil {
		return nil, err
	}

	chances := make([]process.SelectionChance, 0)
	for _, chance := range ratingsConfig.General.SelectionChances {
		chances = append(chances, &SelectionChance{
			MaxThreshold:  chance.MaxThreshold,
			ChancePercent: chance.ChancePercent,
		})
	}

	// avoid any invalid configuration where ratings are not sorted by epoch
	slices.SortFunc(ratingsConfig.ShardChain.RatingStepsByEpoch, func(a, b config.RatingSteps) int {
		return int(a.EnableEpoch) - int(b.EnableEpoch)
	})

	if !checkForEpochZeroConfigurationInSovereign(args) {
		return nil, process.ErrMissingConfigurationForEpochZero
	}

	currentChainParameters := args.ChainParametersHolder.CurrentChainParameters()
	shardRatingStep, err := createShardRatingStep(args, currentChainParameters)
	if err != nil {
		return nil, err
	}

	ratingsConfigValue := ratingsStepsData{
		enableEpoch:          args.EpochNotifier.CurrentEpoch(),
		shardRatingsStepData: shardRatingStep,
		metaRatingsStepData:  shardRatingStep, // filling it so that no nil pointer is used further in code
	}

	ratingData := &RatingsData{
		startRating:                 ratingsConfig.General.StartRating,
		maxRating:                   ratingsConfig.General.MaxRating,
		minRating:                   ratingsConfig.General.MinRating,
		signedBlocksThreshold:       ratingsConfig.General.SignedBlocksThreshold,
		currentRatingsStepData:      ratingsConfigValue,
		selectionChances:            chances,
		chainParametersHandler:      args.ChainParametersHolder,
		ratingsSetup:                ratingsConfig,
		roundDurationInMilliseconds: args.RoundDurationMilliseconds,
		statusHandler:               statusHandler.NewNilStatusHandler(),
	}

	err = ratingData.computeRatingStepsConfig(args.ChainParametersHolder.AllChainParameters(), false)
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(ratingData)

	return &sovereignRatingsData{
		ratingData,
	}, nil
}

func verifySovereignRatingsConfig(settings config.RatingsConfig) error {
	err := verifyGeneralConfig(settings.General)
	if err != nil {
		return err
	}

	return checkRatingStepsByEpochConfigForDest(settings.ShardChain.RatingStepsByEpoch, "sovereignShardChain")
}

func checkForEpochZeroConfigurationInSovereign(args RatingsDataArg) bool {
	_, foundShardChainRatingSteps := getRatingStepsForEpoch(0, args.Config.ShardChain.RatingStepsByEpoch)
	_, foundChainParams := getChainParamsForEpoch(0, args.ChainParametersHolder.AllChainParameters())

	return foundShardChainRatingSteps && foundChainParams
}
