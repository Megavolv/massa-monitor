package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Massa struct {
	logger                         *log.Logger
	PrivateKey, PublicKey, Address string
	SequentialBalance              *SequentialBalance
	ParallelBalance                *ParallelBalance
	Rolls                          *Rolls
}

type SequentialBalance struct {
	Balance decimal.Decimal
}

type ParallelBalance struct {
	Final, Candidate, Locked decimal.Decimal
}

type Rolls struct {
	Active, Candidate, Final decimal.Decimal
}

func NewMassa(log *log.Logger) (m *Massa) {
	m = &Massa{logger: log,
		SequentialBalance: &SequentialBalance{},
		ParallelBalance:   &ParallelBalance{},
		Rolls:             &Rolls{},
	}
	return
}

func (m *Massa) CheckExecutable() (err error) {
	file, err := os.Open("./massa-client")
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("Massa executable is not found")
	}
	return file.Close()
}

func (m *Massa) IsWalletLoaded() error {
	if len(m.PrivateKey) > 0 &&
		len(m.PublicKey) > 0 &&
		len(m.Address) > 0 {
		return nil
	}
	return errors.New("Wallet loading serror")
}

func (m *Massa) Parse(data []string) (err error) {
	m.PrivateKey, err = space_extract(data[1])
	if err != nil {
		return
	}

	m.PublicKey, err = space_extract(data[2])
	if err != nil {
		return
	}

	m.Address, err = space_extract(data[3])
	if err != nil {
		return
	}

	m.SequentialBalance.Balance, err = space_extract_dec(data[6])
	if err != nil {
		return
	}

	m.ParallelBalance.Final, err = space_extract_dec(data[9])
	if err != nil {
		return
	}

	m.ParallelBalance.Candidate, err = space_extract_dec(data[10])
	if err != nil {
		return
	}

	m.ParallelBalance.Locked, err = space_extract_dec(data[11])
	if err != nil {
		return
	}

	m.Rolls.Active, err = space_extract_dec(data[14])
	if err != nil {
		return
	}

	m.Rolls.Final, err = space_extract_dec(data[15])
	if err != nil {
		return
	}

	m.Rolls.Candidate, err = space_extract_dec(data[16])

	return
}

func space_extract(s string) (op string, err error) {
	data := strings.Split(s, " ")
	op = data[len(data)-1]
	return
}

func space_extract_dec(s string) (op decimal.Decimal, err error) {
	data, err := space_extract(s)
	if err != nil {
		return
	}
	return decimal.NewFromString(data)
}

func (m *Massa) Exec(opts []string) ([]byte, error) {
	r := exec.Command("./massa-client", opts...)
	return r.Output()
}

func (m *Massa) LoadWallet() (err error) {
	m.logger.Trace("Load wallet\n")
	d, err := m.Exec([]string{"wallet_info"})
	data := strings.Split(string(d), "\n")

	err = m.Parse(data)

	if err == nil {
		return m.IsWalletLoaded()
	}

	return

}

func (m *Massa) CheckAndStakeKey() (err error) {
	m.logger.Trace("Check stake key\n")
	d, err := m.Exec([]string{"node_get_staking_addresses"})
	if err != nil {
		return err
	}
	data := strings.Split(string(d), "\n")
	if data[0] != m.Address {
		m.logger.Trace("Need to stake\n")
		err = m.RegisterStakeKey()
	} else {
		m.logger.Trace("Ok\n")
	}
	return
}

func (m *Massa) NeedToBuy() (need bool) {
	if m.ParallelBalance.Candidate.IsZero() { // m.ActiveRolls.IsZero()
		need = true
	}

	return
}

func (m *Massa) BuyRolls() (err error) {
	m.logger.Info("Try to buy\n")
	data, err := m.Exec([]string{"buy_rolls", m.Address, "1", "0"})
	if err == nil {
		m.logger.Debug(string(data))
	}
	return
}

func (m *Massa) RegisterStakeKey() (err error) {
	m.logger.Info("Try to stake\n")
	data, err := m.Exec([]string{"node_add_staking_private_keys", m.PrivateKey})
	if err == nil {
		m.logger.Debug(string(data))
	}
	return
}

func (m *Massa) Process() {
	err := m.LoadWallet()
	if err != nil {
		m.logger.Error(err)
		return
	}

	empJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		m.logger.Error(err.Error())
	}

	fmt.Println(string(empJSON))

	if m.NeedToBuy() && m.ParallelBalance.Final.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		if err = m.BuyRolls(); err != nil {
			m.logger.Error(err)
		}
	}

	err = m.CheckAndStakeKey()
	if err != nil {
		m.logger.Error(err)
	}

	return
}
