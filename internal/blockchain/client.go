package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)
// encapsula a conexão com a blockchain
// e os dados necessários para enviar transações.
type BlockchainClient struct {
	Client      *ethclient.Client
	PrivateKey  *ecdsa.PrivateKey
	PublicKey   common.Address
	ChainID     *big.Int
	nonce       uint64
	nonceLoaded bool
}
// representa os parâmetros para inicializar o cliente da blockchain
type Config struct {
	RPC       string
	PrivateKey string
	ChainID   int64
}

func NewBlockchainClient(cfg Config) (*BlockchainClient, error) {
	client, err := ethclient.Dial(cfg.RPC)
	if err != nil {
		return nil, fmt.Errorf("erro RPC: %w", err)
	}

	pk, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key inválida: %w", err)
	}

	pub := crypto.PubkeyToAddress(pk.PublicKey)

	chainID := big.NewInt(cfg.ChainID)
	
	nonce, err := client.PendingNonceAt(context.Background(), pub)
	if err != nil {
		return nil, fmt.Errorf("erro pegando nonce: %w", err)
	}

	log.Println("Blockchain conectado. Conta:", pub.Hex())

	return &BlockchainClient{
		Client:      client,
		PrivateKey:  pk,
		PublicKey:   pub,
		ChainID:     chainID,
		nonce:       nonce,
		nonceLoaded: true,
	}, nil
}
//  cria as opções de transação (TransactOpts)
// para enviar transações assinadas
func (bc *BlockchainClient) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	myNonce := bc.nonce
	bc.nonce++ // controla o nonce para n ter concorrencia de numero da transação 

	gasPrice, err := bc.Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(bc.PrivateKey, bc.ChainID)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(myNonce))
	auth.GasLimit = 3_000_000
	auth.GasPrice = gasPrice

	return auth, nil
}

//  retorna opções para chamadas de leitura (call),
// não enviam transação, não gastam gas
func (bc *BlockchainClient) CallOpts() *bind.CallOpts {
	return &bind.CallOpts{
		From: bc.PublicKey,
	}
}

// Ping para verificar conexão
func (bc *BlockchainClient) Ping(ctx context.Context) error {
	_, err := bc.Client.BlockNumber(ctx)
	return err
}