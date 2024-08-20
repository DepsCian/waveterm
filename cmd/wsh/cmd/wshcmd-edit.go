// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/thenextwave/pkg/waveobj"
	"github.com/wavetermdev/thenextwave/pkg/wshrpc"
	"github.com/wavetermdev/thenextwave/pkg/wshrpc/wshclient"
)

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "edit a file",
	Args:    cobra.ExactArgs(1),
	Run:     editRun,
	PreRunE: preRunSetupRpcClient,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func editRun(cmd *cobra.Command, args []string) {
	fileArg := args[0]
	wshCmd := wshrpc.CommandCreateBlockData{
		BlockDef: &waveobj.BlockDef{
			Meta: map[string]any{
				waveobj.MetaKey_View: "preview",
				waveobj.MetaKey_File: fileArg,
			},
		},
	}
	if RpcContext.Conn != "" {
		wshCmd.BlockDef.Meta[waveobj.MetaKey_Connection] = RpcContext.Conn
	}
	absFile, err := filepath.Abs(fileArg)
	if err != nil {
		WriteStderr("[error] getting absolute path: %v\n", err)
		return
	}
	_, err = os.Stat(absFile)
	if err == fs.ErrNotExist {
		WriteStderr("[error] file does not exist: %q\n", absFile)
		return
	}
	if err != nil {
		WriteStderr("[error] getting file info: %v\n", err)
		return
	}
	blockRef, err := wshclient.CreateBlockCommand(RpcClient, wshCmd, &wshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		WriteStderr("[error] running view command: %v\r\n", err)
		return
	}
	doneCh := make(chan bool)
	RpcClient.EventListener.On("blockclose", func(event *wshrpc.WaveEvent) {
		if event.HasScope(blockRef.String()) {
			close(doneCh)
		}
	})
	wshclient.EventSubCommand(RpcClient, wshrpc.SubscriptionRequest{Event: "blockclose", Scopes: []string{blockRef.String()}}, nil)
	<-doneCh
}
