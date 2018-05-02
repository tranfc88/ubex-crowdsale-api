package crowdsale

import (
    "ubex-api/solidity/bindings/ubex_crowdsale"
    "github.com/ethereum/go-ethereum/common"
    "github.com/spf13/viper"
    "errors"
    "fmt"
    "math/big"
    "ubex-api/common/ethereum"
    "ubex-api/models"
    modelsCommon "ubex-api/common/models"
    "github.com/ethereum/go-ethereum/core/types"
    "ubex-api/token"
)

var cr *Crowdsale

type Crowdsale struct {
    *ethereum.Contract
    Crowdsale *ubex_crowdsale.UbexCrowdsale
}

func Init() error {
    c := ethereum.NewContract(viper.GetString("ethereum.address.crowdsale"))
    c.InitEvents(ubex_crowdsale.UbexCrowdsaleABI)

    s, err := ubex_crowdsale.NewUbexCrowdsale(c.Address, c.Wallet.Connection)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to instantiate a Crowdsale contract: %v", err))
    }

    cr = &Crowdsale{
        Contract: c,
        Crowdsale: s,
    }

    return nil
}

func GetCrowdsale() *Crowdsale {
    return cr
}

func (s *Crowdsale) Deploy(params *models.CrowdsaleDeployParams) (*common.Address, *types.Transaction, error) {
    tokenAddr := token.GetToken().Address

    tokenRate, ok := big.NewInt(0).SetString(params.TokenRate, 0)
    if !ok {
        return nil, nil, fmt.Errorf("wrong TokenRate provided: %s", params.TokenRate)
    }

    address, tx, _, err := ubex_crowdsale.DeployUbexCrowdsale(
        s.Wallet.Account,
        s.Wallet.Connection,
        tokenRate,
        common.HexToAddress(params.WalletAddress),
        tokenAddr,
    )
    if err != nil {
        return nil, nil, fmt.Errorf("failed to deploy contract: %v", err)
    }
    return &address, tx, nil
}

func (s *Crowdsale) Balance(addr string) (*big.Int, error) {
    return s.Crowdsale.Balances(nil, common.HexToAddress(addr))
}

func (s *Crowdsale) Status() (*models.CrowdsaleStatus, error) {
    weiRaised, err := s.Crowdsale.WeiRaised(nil)
    if err != nil {
        return nil, err
    }

    rate, err := s.Crowdsale.Rate(nil)
    if err != nil {
        return nil, err
    }

    tokensIssued, err := s.Crowdsale.TokensIssued(nil)
    if err != nil {
        return nil, err
    }

    return &models.CrowdsaleStatus{
        Address: s.Address.String(),
        TokensIssued: tokensIssued.String(),
        Rate: rate.String(),
        WeiRaised: weiRaised.String(),
    }, nil
}

func (s *Crowdsale) Events(addrs []string) ([]modelsCommon.ContractEvent, error) {
    hashAddrs := make([]common.Hash, len(addrs))
    for _, addr := range addrs {
        hashAddrs = append(hashAddrs, common.HexToHash(addr))
    }

    events, err := s.GetEventsByTopics(
        [][]common.Hash{{}, hashAddrs},
        big.NewInt(viper.GetInt64("ethereum.start_block.crowdsale")),
    )
    if err != nil {
        return nil, err
    }

    resEvents := make([]modelsCommon.ContractEvent, 0)

    for _, event := range events {
        switch {
        case event.Name == "TokenPaid":
            event.Args = models.TokenPaidEventArgs{
                Purchaser: common.BytesToAddress(event.RawArgs[0]).String(),
                Beneficiary: common.BytesToAddress(event.RawArgs[1]).String(),
                WeiAmount: common.BytesToHash(event.RawArgs[2]).Big().String(),
                Created: common.BytesToHash(event.RawArgs[3]).Big().String(),
            }
        case event.Name == "TokenPurchase":
            event.Args = models.TokenPurchaseEventArgs{
                Purchaser: common.BytesToAddress(event.RawArgs[0]).String(),
                Beneficiary: common.BytesToAddress(event.RawArgs[1]).String(),
                WeiAmount: common.BytesToHash(event.RawArgs[2]).Big().String(),
                TokensAmount: common.BytesToHash(event.RawArgs[3]).Big().String(),
            }
        default:
            return nil, fmt.Errorf("unknown event type: %s", event.Name)
        }

        resEvents = append(resEvents, event)
    }

    return resEvents, nil
}