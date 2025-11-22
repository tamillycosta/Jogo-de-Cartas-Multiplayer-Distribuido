package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)
// Estrutura para repsentar um cliente conectado a blockchain
type BlockchainClient struct {
	Client      *ethclient.Client
	PrivateKey  *ecdsa.PrivateKey
	PublicKey   common.Address
	ChainID     *big.Int
	
	nonce      uint64
	nonceMutex sync.Mutex
	
	playerNonces map[string]uint64
	playerMutex  sync.Mutex
}

type Config struct {
	RPC        string
	PrivateKey string
	ChainID    int64
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

	log.Printf(" Blockchain conectado. Conta: %s, Nonce inicial: %d", pub.Hex(), nonce)

	return &BlockchainClient{
		Client:       client,
		PrivateKey:   pk,
		PublicKey:    pub,
		ChainID:      chainID,
		nonce:        nonce,
		playerNonces: make(map[string]uint64),
	}, nil
}

// ===== TRANSACTOR DO SERVIDOR =====

func (bc *BlockchainClient) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	bc.nonceMutex.Lock()
	defer bc.nonceMutex.Unlock()

	myNonce := bc.nonce
	bc.nonce++

	log.Printf(" [Nonce] Servidor usando nonce: %d", myNonce)

	gasPrice, err := bc.Client.SuggestGasPrice(ctx)
	if err != nil {
		bc.nonce--
		return nil, fmt.Errorf("erro ao obter gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(bc.PrivateKey, bc.ChainID)
	if err != nil {
		bc.nonce--
		return nil, fmt.Errorf("erro ao criar transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(myNonce))
	auth.GasLimit = 3_000_000
	auth.GasPrice = gasPrice
	auth.Context = ctx

	return auth, nil
}

// ===== TRANSACTOR DO JOGADOR =====

func (bc *BlockchainClient) NewTransactorFromPrivateKey(ctx context.Context, hexKey string) (*bind.TransactOpts, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")

	pk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("private key inválida: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(pk.PublicKey)
	addrStr := fromAddr.Hex()

	bc.playerMutex.Lock()
	
	if _, exists := bc.playerNonces[addrStr]; !exists {
		nonce, err := bc.Client.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			bc.playerMutex.Unlock()
			return nil, fmt.Errorf("erro pegando nonce do jogador: %w", err)
		}
		bc.playerNonces[addrStr] = nonce
		log.Printf(" [Nonce] Jogador %s nonce inicial: %d", addrStr, nonce)
	}

	myNonce := bc.playerNonces[addrStr]
	bc.playerNonces[addrStr]++
	bc.playerMutex.Unlock()

	log.Printf(" [Nonce] Jogador %s usando nonce: %d", addrStr, myNonce)

	gasPrice, err := bc.Client.SuggestGasPrice(ctx)
	if err != nil {
		bc.playerMutex.Lock()
		bc.playerNonces[addrStr]--
		bc.playerMutex.Unlock()
		return nil, fmt.Errorf("erro ao obter gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, bc.ChainID)
	if err != nil {
		bc.playerMutex.Lock()
		bc.playerNonces[addrStr]--
		bc.playerMutex.Unlock()
		return nil, fmt.Errorf("erro ao criar transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(myNonce))
	auth.GasPrice = gasPrice
	auth.GasLimit = 3_000_000
	auth.Context = ctx

	return auth, nil
}

// ===== FAUCET: ENVIAR ETH PARA JOGADOR =====

func (bc *BlockchainClient) FundAccount(ctx context.Context, toAddress string, amountWei *big.Int) error {
	bc.nonceMutex.Lock()
	myNonce := bc.nonce
	bc.nonce++
	bc.nonceMutex.Unlock()

	toAddr := common.HexToAddress(toAddress)

	gasPrice, err := bc.Client.SuggestGasPrice(ctx)
	if err != nil {
		bc.nonceMutex.Lock()
		bc.nonce--
		bc.nonceMutex.Unlock()
		return fmt.Errorf("erro ao obter gas price: %w", err)
	}

	// Criar transação de transferência de ETH
	tx := types.NewTransaction(
		myNonce,
		toAddr,
		amountWei,
		21000,     // Gas limit padrão para transferência simples
		gasPrice,
		nil,       // Sem data (transferência simples)
	)

	// Assinar transação
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bc.ChainID), bc.PrivateKey)
	if err != nil {
		bc.nonceMutex.Lock()
		bc.nonce--
		bc.nonceMutex.Unlock()
		return fmt.Errorf("erro ao assinar transação: %w", err)
	}

	// Enviar transação
	err = bc.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		bc.nonceMutex.Lock()
		bc.nonce--
		bc.nonceMutex.Unlock()
		return fmt.Errorf("erro ao enviar transação: %w", err)
	}

	log.Printf(" [Faucet] Enviando %s wei para %s. TX: %s", 
		amountWei.String(), toAddress, signedTx.Hash().Hex())

	// Aguardar confirmação
	receipt, err := bind.WaitMined(ctx, bc.Client, signedTx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	log.Printf(" [Faucet] Transferência confirmada no bloco %d", receipt.BlockNumber.Uint64())

	return nil
}

// ===== CONSULTAR SALDO =====

func (bc *BlockchainClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := bc.Client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar saldo: %w", err)
	}
	return balance, nil
}

// ===== HELPERS =====

func (bc *BlockchainClient) ResyncNonce(ctx context.Context) error {
	bc.nonceMutex.Lock()
	defer bc.nonceMutex.Unlock()

	nonce, err := bc.Client.PendingNonceAt(ctx, bc.PublicKey)
	if err != nil {
		return fmt.Errorf("erro ao resync nonce: %w", err)
	}

	oldNonce := bc.nonce
	bc.nonce = nonce
	log.Printf(" [Nonce] Resync servidor: %d -> %d", oldNonce, nonce)

	return nil
}

func (bc *BlockchainClient) ResyncPlayerNonce(ctx context.Context, playerAddress string) error {
	bc.playerMutex.Lock()
	defer bc.playerMutex.Unlock()

	addr := common.HexToAddress(playerAddress)
	nonce, err := bc.Client.PendingNonceAt(ctx, addr)
	if err != nil {
		return fmt.Errorf("erro ao resync nonce do jogador: %w", err)
	}

	oldNonce := bc.playerNonces[playerAddress]
	bc.playerNonces[playerAddress] = nonce
	log.Printf(" [Nonce] Resync jogador %s: %d -> %d", playerAddress, oldNonce, nonce)

	return nil
}

func (bc *BlockchainClient) ClearPlayerNonce(playerAddress string) {
	bc.playerMutex.Lock()
	defer bc.playerMutex.Unlock()
	delete(bc.playerNonces, playerAddress)
}

func (bc *BlockchainClient) CallOpts() *bind.CallOpts {
	return &bind.CallOpts{
		From: bc.PublicKey,
	}
}

func (bc *BlockchainClient) Ping(ctx context.Context) error {
	_, err := bc.Client.BlockNumber(ctx)
	return err
}

// ===== CONVERTER ETH PARA WEI =====

func EthToWei(eth float64) *big.Int {
	// 1 ETH = 10^18 Wei
	weiPerEth := new(big.Float).SetFloat64(1e18)
	ethFloat := new(big.Float).SetFloat64(eth)
	
	weiFloat := new(big.Float).Mul(ethFloat, weiPerEth)
	
	wei := new(big.Int)
	weiFloat.Int(wei)
	return wei
}

func WeiToEth(wei *big.Int) float64 {
	weiFloat := new(big.Float).SetInt(wei)
	ethFloat := new(big.Float).Quo(weiFloat, new(big.Float).SetFloat64(1e18))
	
	eth, _ := ethFloat.Float64()
	return eth
}