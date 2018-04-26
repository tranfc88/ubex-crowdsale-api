package token

import (
    "ubex-api/solidity/bindings/ubex_token"
    "github.com/ethereum/go-ethereum/common"
    "github.com/spf13/viper"
    "errors"
    "fmt"
    "math/big"
    "ubex-api/common/ethereum"
    "github.com/ethereum/go-ethereum/core/types"
)

var token *Token

type Token struct {
    *ethereum.Contract
    Token *ubex_token.UbexToken
}

func Init() error {
    c := ethereum.NewContract(viper.GetString("ethereum.address.token"))
    c.InitEvents(ubex_token.UbexTokenABI)

    t, err := ubex_token.NewUbexToken(c.Address, c.Wallet.Connection)
    if err != nil {
        return errors.New(fmt.Sprintf("Failed to instantiate a Token contract: %v", err))
    }

    token = &Token{
        Contract: c,
        Token:    t,
    }

    return nil
}

func (t *Token) Deploy(totSupply string) (*common.Address, *types.Transaction, error) {
    totSupplyEth, ok := big.NewInt(0).SetString(totSupply, 0)
    if !ok {
        return nil, nil, fmt.Errorf("wrong number provided: %s", totSupply)
    }

    address, tx, _, err := ubex_token.DeployUbexToken(
        t.Wallet.Account,
        t.Wallet.Connection,
        totSupplyEth,
    )
    if err != nil {
        return nil, nil, fmt.Errorf("failed to deploy UbexToken contract: %v", err)
    }
    return &address, tx, nil
}

func GetToken() *Token {
    return token
}

func (t *Token) Balance(addr string) (*big.Int, error) {
    return t.Token.BalanceOf(nil, common.HexToAddress(addr))
}
