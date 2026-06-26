package rpc

import (
	"chronix/cmd/app/consts"
	"chronix/pkg/fileutil"
	"fmt"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"
)

var (
	listener   net.Listener
	SocketPath string
)

func DefaultSocketPath() string {
	if p := os.Getenv("RPC_SOCKET_PATH"); p != "" {
		return p
	}

	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		return filepath.Join(programData, "Chronix", consts.APPNAME+"-rpc.sock")
	}

	if runtime.GOOS == "darwin" && os.Getuid() == 0 {
		return filepath.Join("/Library/Application Support", "Chronix", consts.APPNAME+"-rpc.sock")
	}

	if runtime.GOOS == "linux" && os.Getuid() == 0 {
		return filepath.Join("/run", consts.APPNAME, consts.APPNAME+"-rpc.sock")
	}

	if r := os.Getenv("XDG_RUNTIME_DIR"); r != "" {
		if info, err := os.Stat(r); err == nil && info.IsDir() {
			// On Unix, verify that the directory belongs to us.
			// XDG_RUNTIME_DIR MUST be owned by the user.
			if !fileutil.IsOwnedByUser(info) {
				goto fallback
			}
			return filepath.Join(r, consts.APPNAME, consts.APPNAME+"-rpc.sock")
		}
	}

fallback:
	return filepath.Join(os.TempDir(), consts.APPNAME+"-rpc.sock")
}

func init() {
	SocketPath = DefaultSocketPath()
}

func RegisterName(name string, rcvr any) error {
	if err := rpc.RegisterName(name, rcvr); err != nil {
		slog.Error("rpc register name", "error", err)
		return err
	}
	return nil
}

func Shutdown() {
	if listener != nil {
		_ = listener.Close()
	}
	_ = os.Remove(SocketPath)
}

func StartServer() error {
	_ = os.Remove(SocketPath)
	if err := os.MkdirAll(filepath.Dir(SocketPath), 0o770); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}
	var err error
	listener, err = net.Listen("unix", SocketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	slog.Info("RPC server listening", "path", SocketPath)
	_ = os.Chmod(SocketPath, 0o660)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if _, ok := err.(*net.OpError); ok {
					return
				}
				slog.Error("rpc accept", "error", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
	return nil
}

func Client() (*rpc.Client, error) {
	client, err := rpc.Dial("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("daemon not found (could not connect to %s): %w", SocketPath, err)
	}
	return client, nil
}

func Call(serviceMethod string, args any, reply any) error {
	if args == nil {
		args = &struct{}{}
	}
	if reply == nil {
		reply = &struct{}{}
	}
	cl, err := Client()
	if err != nil {
		return err
	}
	defer func() { _ = cl.Close() }()
	return cl.Call(serviceMethod, args, reply)
}
