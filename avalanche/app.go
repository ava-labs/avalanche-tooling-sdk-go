// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import "github.com/ava-labs/avalanchego/utils/logging"

type Avalanche struct {
	Log logging.Logger
	//baseDir string
	//Conf       *config.Config
	//Prompt     prompts.Prompter
	//Apm        *apm.APM
	//ApmDir     string
	//Downloader Downloader
}

func New() *Avalanche {
	return &Avalanche{}
}

func (app *Avalanche) Setup(log logging.Logger) {
	//app.baseDir = baseDir
	app.Log = log
	//app.Conf = conf
	//app.Prompt = prompt
	//app.Downloader = downloader
}
