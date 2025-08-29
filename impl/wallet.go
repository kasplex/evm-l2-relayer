package impl

import (
	"encoding/hex"
	"fmt"

	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/utils"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kasplex-evm/kasplex-relayer/log"
)

const (
	U64MaxValue  = 18446744073709551615
	BaseFee      = "0.0003"
	CompressSize = 2048
)

type Wallet struct {
	pk               string
	keyPair          *secp256k1.SchnorrKeyPair
	schnorrPublicKey *secp256k1.SchnorrPublicKey
	addr             *util.AddressPublicKey
	url              string
	clientPool       *RPCClientPool
}

func NewWallet(pk string, url string) (w *Wallet, err error) {
	keyBytes, err := hex.DecodeString(pk)
	if err != nil {
		return nil, err
	}

	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(keyBytes)
	if err != nil {
		return nil, err
	}

	schnorrPublicKey, err := keyPair.SchnorrPublicKey()
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := hex.DecodeString(schnorrPublicKey.String())
	if err != nil {
		return nil, err
	}

	addr, err := util.NewAddressPublicKey(publicKeyBytes, util.Bech32PrefixKaspaTest)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		pk:               pk,
		keyPair:          keyPair,
		schnorrPublicKey: schnorrPublicKey,
		addr:             addr,
		url:              url,
		clientPool:       NewRPCClientPool(url, DefaultPoolSize),
	}, nil
}

func (w *Wallet) TransferVM(
	toAddr string,
	kas string,
	vmData []byte,
	isJSON bool,
) (string, error) {
	client, err := w.clientPool.getClient()
	if err != nil {
		return "", err
	}
	defer w.clientPool.putClient(client)

	tx := new(externalapi.DomainTransaction)
	tx.Inputs = make([]*externalapi.DomainTransactionInput, 0)

	uxto, err := client.GetUTXOsByAddresses([]string{w.addr.String()})
	if err != nil {
		return "", err
	}

	totalAmount := uint64(0)
	targetAmount, err := utils.KasToSompi(kas)
	if err != nil {
		return "", err
	}

	fee, err := utils.KasToSompi(BaseFee)
	if err != nil {
		return "", err
	}

	for _, v := range uxto.Entries {
		if totalAmount >= targetAmount+fee {
			break
		}
		op, err := appmessage.RPCOutpointToDomainOutpoint(v.Outpoint)
		if err != nil {
			continue
		}
		entry, err := appmessage.RPCUTXOEntryToUTXOEntry(v.UTXOEntry)
		if err != nil {
			continue
		}
		tx.Inputs = append(tx.Inputs, &externalapi.DomainTransactionInput{
			PreviousOutpoint: *op,
			SignatureScript:  []byte{},
			Sequence:         U64MaxValue,
			UTXOEntry:        entry,
			SigOpCount:       1,
		})
		totalAmount += v.UTXOEntry.Amount
	}

	if totalAmount < targetAmount+fee {
		return "", fmt.Errorf("insufficient balance")
	}

	to, err := util.DecodeAddress(toAddr, util.Bech32PrefixKaspaTest)
	if err != nil {
		log.Infof("util.DecodeAddress error:", err)
		return "", err
	}
	payToAddr, _ := txscript.PayToAddrScript(to)
	tx.Outputs = []*externalapi.DomainTransactionOutput{
		{
			Value: targetAmount,
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  payToAddr.Script,
				Version: 0,
			},
		},
	}
	change := totalAmount - targetAmount - fee
	if change > 0 {
		changeAddr, _ := util.DecodeAddress(w.addr.String(), util.Bech32PrefixKaspaTest)
		tmp1, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			return "", err
		}
		tx.Outputs = append(tx.Outputs, &externalapi.DomainTransactionOutput{
			Value: change,
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  tmp1.Script,
				Version: 0,
			},
		})
	}

	tx.Payload = []byte("kasplex")
	isCompressed := false
	if len(vmData) > CompressSize {
		compressed, err := zlibCompress(vmData)
		if err != nil {
			log.Infof("zlibCompress error:", err)
		} else {
			isCompressed = true
			log.Infof("Data compressed, size:", len(vmData), " ->", len(compressed))
			vmData = compressed

		}
	}
	if isCompressed {
		if isJSON {
			tx.Payload = append(tx.Payload, 0x80)
		} else {
			tx.Payload = append(tx.Payload, 0x81)
		}
	} else {
		if isJSON {
			tx.Payload = append(tx.Payload, 0x0)
		} else {
			tx.Payload = append(tx.Payload, 0x1)
		}
	}
	tx.Payload = append(tx.Payload, vmData...)

	sighash := new(consensushashing.SighashReusedValues)
	for i := range tx.Inputs {
		domainHash, err := consensushashing.CalculateSignatureHashSchnorr(tx, i, consensushashing.SigHashAll, sighash)
		if err != nil {
			return "", err
		}
		hash := new(secp256k1.Hash)
		arr := domainHash.ByteSlice()
		hash.SetBytes(arr)
		sign, _ := w.keyPair.SchnorrSign(hash)
		signBytes, _ := hex.DecodeString(sign.String())
		signBytes = append(signBytes, byte(consensushashing.SigHashAll))
		sb := txscript.NewScriptBuilder()
		sb.AddData(signBytes)
		scriptBytes, _ := sb.Script()
		tx.Inputs[i].SignatureScript = scriptBytes
	}

	rpcTx := appmessage.DomainTransactionToRPCTransaction(tx)
	res, err := client.SubmitTransaction(rpcTx, consensushashing.TransactionID(tx).String(), false)
	if err != nil {
		return "", err
	}

	log.Infof("Transaction submitted:", res.TransactionID)
	return res.TransactionID, nil
}

func (w *Wallet) Close() {
	if w.clientPool != nil {
		w.clientPool.Close()
	}
}
