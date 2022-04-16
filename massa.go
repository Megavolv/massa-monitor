package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
	Candidate decimal.Decimal
	Final     decimal.Decimal
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
	m.logger.Info(data)

	if len(data) < 18 {
		return errors.New("len(data)=" + strconv.Itoa(len(data)))
	}

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

	/*	m.SequentialBalance.Final, err = space_extract_dec(data[6])
		if err != nil {
			return
		}*/

	m.logger.Trace("Parse m.ParallelBalance.Final", data[12])
	m.ParallelBalance.Final, err = space_extract_dec(data[12])
	if err != nil {
		return
	}

	/*	m.ParallelBalance.Candidate, err = space_extract_dec(data[10])
		if err != nil {
			return
		}

		m.ParallelBalance.Locked, err = space_extract_dec(data[11])
		if err != nil {
			return
		}
	*/
	m.logger.Trace("Parse m.Rolls.Active", data[17])
	m.Rolls.Active, err = space_extract_dec(data[17])
	if err != nil {
		return
	}

	m.logger.Trace("Parse m.Rolls.Final", data[18])
	m.Rolls.Final, err = space_extract_dec(data[18])
	if err != nil {
		return
	}

	m.logger.Trace("Parse m.Rolls.Candidate", data[19])
	m.Rolls.Candidate, err = space_extract_dec(data[19])

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
		err = errors.New(err.Error() + " on " + s)
		return
	}
	op, err = decimal.NewFromString(data)
	if err != nil {
		err = errors.New(err.Error() + " on " + s)
	}

	return
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
	return m.Rolls.Candidate.IsZero()
}

func (m *Massa) BuyRolls() (err error) {
	m.logger.Warn("Try to buy\n")
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
		return
	}

	fmt.Println(string(empJSON))

	m.logger.Info("ParallelBalance.Final:", m.ParallelBalance.Final.String())

	if m.NeedToBuy() && m.ParallelBalance.Final.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		if err = m.BuyRolls(); err != nil {
			m.logger.Error(err)
		} else {
		}
	} else {

	}

	err = m.CheckAndStakeKey()
	if err != nil {
		m.logger.Error(err)
	}

	return
}
