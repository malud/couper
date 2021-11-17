package command

import (
	"fmt"

	"github.com/avenga/couper/config"
	"github.com/avenga/couper/config/configload"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
)

var _ Cmd = &Verify{}

type Verify struct{}

func NewVerify() *Verify {
	return &Verify{}
}

func (v Verify) Execute(args Args, _ *config.Couper, logger *logrus.Entry) error {
	if len(args) != 1 {
		v.Usage()

		err := fmt.Errorf("invalid number of arguments given")
		logger.WithError(err).Error()

		return err
	}

	_, err := configload.LoadFile(args[0], true)
	if diags, ok := err.(hcl.Diagnostics); ok {
		for _, diag := range diags {
			logger.WithError(diag).Error()
		}

		return diags
	}

	return err
}

func (v Verify) Usage() {
	println("Usage of verify:\n  verify [-f <file>]	Verify the syntax of the given configuration file.")
}