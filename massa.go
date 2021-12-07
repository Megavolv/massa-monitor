package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	//"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Massa struct {
	logger                                    *log.Logger
	PrivateKey, PublicKey, Address            string
	FinalBalance, ActiveRolls, CandidateRolls decimal.Decimal
}

func NewMassa(log *log.Logger) (m *Massa) {
	m = &Massa{logger: log}
	return
}

func (m *Massa) CheckExecutable() (err error) {
	file, err := os.Open("./massa-client")
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("Massa executable is not found")
	}
	return file.Close()
}

func (m *Massa) Parse(data []string) (err error) {
	m.PrivateKey, err = space_extract(data[1], 2)
	if err != nil {
		return
	}

	m.PublicKey, err = space_extract(data[2], 2)
	if err != nil {
		return
	}

	m.Address, err = space_extract(data[3], 1)
	if err != nil {
		return
	}

	m.FinalBalance, err = space_extract_dec(data[6], 2)
	if err != nil {
		return
	}

	m.ActiveRolls, err = space_extract_dec(data[11], 2)
	if err != nil {
		return
	}

	m.CandidateRolls, err = space_extract_dec(data[13], 2)
	if err != nil {
		return
	}

	return
}

func space_extract(s string, num int) (op string, err error) {
	data := strings.Split(s, " ")
	if len(data) < num {
		err = errors.New("Cannot parse")
		return
	}
	op = data[num]
	return
}

func space_extract_dec(s string, num int) (op decimal.Decimal, err error) {
	data, err := space_extract(s, num)
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
	if len(data) != 22 {
		m.logger.Debug(data)
		err = errors.New("Not 22 lines in output")
		return
	}

	err = m.Parse(data)

	return

}

func (m *Massa) NeedToBuy() (need bool) {
	if m.CandidateRolls.IsZero() { // m.ActiveRolls.IsZero()
		need = true
	}

	return
}

func (m *Massa) BuyRolls() (err error) {
	m.logger.Info("Try to buy\n")
	data, err := m.Exec([]string{"buy_rolls", m.Address, "1", "0"})
	if err == nil {
		m.logger.Debug(data)
	}
	return
}

func (m *Massa) RegisterStakeKey() (err error) {
	m.logger.Info("Try to stake\n")
	data, err := m.Exec([]string{"node_add_staking_private_keys", m.PrivateKey})
	if err == nil {
		m.logger.Debug(data)
	}
	return
}

func (m *Massa) Process() {
	err := m.LoadWallet()
	if err != nil {
		m.logger.Error(err)
		return
	}

	fmt.Printf("PrivateKey: %s\n PublicKey: %s\n Address: %s\n FinalBalance: %s\n ActiveRolls: %s\n CandidateRolls: %s\n", m.PrivateKey, m.PublicKey, m.Address, m.FinalBalance.String(), m.ActiveRolls.String(), m.CandidateRolls.String())

	if m.NeedToBuy() && m.FinalBalance.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		if err = m.BuyRolls(); err == nil {
			err = m.RegisterStakeKey()
		}
	}

	/*if err == nil {
		m.logger.Info("No action need")
	}*/

	return
}
