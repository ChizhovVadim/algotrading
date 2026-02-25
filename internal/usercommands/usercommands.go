package usercommands

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

// потом можно прикрутить, чтобы команды не только из консоли, но например из телеграм бота.
// прикрутить, чтобы пользователь мог сам ввести заявку для любого брокера
func Read(
	ctx context.Context,
	messages chan<- any,
) error {
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		var msg, ok = parseUserCmd(commandLine)
		if !ok {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case messages <- msg:
			if _, ok := msg.(model.ExitUserCmd); ok {
				return nil
			}
		}
	}
	return scanner.Err()
}

func parseUserCmd(commandLine string) (any, bool) {
	var tokens = NewTokens(commandLine)
	var commandName = tokens.Next()
	if commandName == "quit" || commandName == "exit" {
		return model.ExitUserCmd{}, true
	}
	if commandName == "status" {
		return model.CheckStatusUserCmd{}, true
	}
	if commandName == "initlimits" {
		var res = model.InitLimitsUserCmd{}
		for {
			var token = tokens.Next()
			if token == "" {
				break
			}
			if token == "client" {
				res.Client = tokens.Next()
			}
		}
		return res, true
	}
	if commandName == "closeall" {
		return model.CloseAllUserCmd{}, true
	}
	return nil, false
}

type Tokens struct {
	fields []string
}

func NewTokens(line string) Tokens {
	return Tokens{fields: strings.Fields(line)}
}

func (t *Tokens) Next() string {
	if len(t.fields) == 0 {
		return ""
	}
	var res = t.fields[0]
	t.fields = t.fields[1:]
	return res
}
