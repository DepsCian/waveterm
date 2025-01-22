// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package connparse

import (
	"context"
	"fmt"
	"strings"

	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshutil"
)

const (
	ConnectionTypeWsh  = "wsh"
	ConnectionTypeS3   = "s3"
	ConnectionTypeWave = "wavefile"

	ConnHostCurrent = "current"
	ConnHostWaveSrv = "wavesrv"
)

type Connection struct {
	Scheme string
	Host   string
	Path   string
}

func (c *Connection) GetSchemeParts() []string {
	return strings.Split(c.Scheme, ":")
}

func (c *Connection) GetType() string {
	lastInd := strings.LastIndex(c.Scheme, ":")
	if lastInd == -1 {
		return c.Scheme
	}
	return c.Scheme[lastInd+1:]
}

func (c *Connection) GetPathWithHost() string {
	if c.Host == "" {
		return ""
	}
	if strings.HasPrefix(c.Path, "/") {
		return c.Host + c.Path
	}
	return c.Host + "/" + c.Path
}

func (c *Connection) GetFullURI() string {
	return c.Scheme + "://" + c.GetPathWithHost()
}

func ParseURIAndReplaceCurrentHost(ctx context.Context, uri string) (*Connection, error) {
	conn, err := ParseURI(uri)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection: %v", err)
	}
	if conn.Host == ConnHostCurrent {
		handler := wshutil.GetRpcResponseHandlerFromContext(context.Background())
		if handler == nil {
			return nil, fmt.Errorf("error getting rpc response handler from context")
		}
		source := handler.GetRpcContext().Conn

		// RPC context connection is empty for local connections
		if source == "" {
			source = wshrpc.LocalConnName
		}
		conn.Host = source
	}
	return conn, nil
}

// ParseURI parses a connection URI and returns the connection type, host/path, and parameters.
func ParseURI(uri string) (*Connection, error) {
	split := strings.SplitN(uri, "://", 2)
	var scheme string
	var rest string
	if len(split) > 1 {
		scheme = split[0]
		rest = split[1]
	} else {
		rest = split[0]
	}

	var host string
	var remotePath string
	if scheme == "" {
		scheme = ConnectionTypeWsh
		if strings.HasPrefix(rest, "//") {
			rest = strings.TrimPrefix(rest, "//")
			split = strings.SplitN(rest, "/", 2)
			if len(split) > 1 {
				host = split[0]
				remotePath = "/" + split[1]
			} else {
				host = split[0]
				remotePath = "/"
			}
		} else if strings.HasPrefix(rest, "/~") {
			host = wshrpc.LocalConnName
			remotePath = rest
		} else {
			host = ConnHostCurrent
			remotePath = rest
		}
	} else {
		split = strings.SplitN(rest, "/", 2)
		if len(split) > 1 {
			host = split[0]
			remotePath = "/" + split[1]
		} else {
			host = split[0]
			remotePath = "/"
		}
	}

	if scheme == ConnectionTypeWsh {
		if host == "" {
			host = wshrpc.LocalConnName
		}
		if strings.HasPrefix(remotePath, "/~") {
			remotePath = strings.TrimPrefix(remotePath, "/")
		}
	}

	conn := &Connection{
		Scheme: scheme,
		Host:   host,
		Path:   remotePath,
	}
	return conn, nil
}
