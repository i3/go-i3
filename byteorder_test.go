package i3

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func msgBytes(order binary.ByteOrder, t messageType, payload string) []byte {
	var buf bytes.Buffer
	if err := binary.Write(&buf, order, &header{magic, uint32(len(payload)), t}); err != nil {
		panic(err)
	}
	_, err := buf.WriteString(payload)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestDetectByteOrder(t *testing.T) {
	t.Parallel()

	for _, i3order := range []binary.ByteOrder{binary.BigEndian, binary.LittleEndian} {
		i3order := i3order // copy
		t.Run(fmt.Sprintf("%T", i3order), func(t *testing.T) {
			t.Parallel()

			var (
				subscribeRequest = msgBytes(i3order, messageTypeSubscribe, "[]"+strings.Repeat(" ", 65536+256-2))
				subscribeReply   = msgBytes(i3order, messageReplyTypeSubscribe, `{"success": true}`)

				nopPrefix         = "nop byte-order detection. padding: "
				runCommandRequest = msgBytes(i3order, messageTypeRunCommand, nopPrefix+strings.Repeat("a", 65536+256-len(nopPrefix)))
				runCommandReply   = msgBytes(i3order, messageReplyTypeCommand, `[{"success": true}]`)

				protocol = map[string][]byte{
					string(subscribeRequest):  subscribeReply,
					string(runCommandRequest): runCommandReply,
				}
			)

			// Abstract socket addresses are a linux-only feature, so we must
			// use file system paths for listening/dialing:
			dir, err := ioutil.TempDir("", "i3test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			path := filepath.Join(dir, fmt.Sprintf("i3test-%T.sock", i3order))
			i3addr, err := net.ResolveUnixAddr("unix", path)
			if err != nil {
				t.Fatal(err)
			}
			i3ln, err := net.ListenUnix("unix", i3addr)
			if err != nil {
				t.Fatal(err)
			}

			var (
				eg       errgroup.Group
				order    binary.ByteOrder
				orderErr error
			)
			eg.Go(func() error {
				addr, err := net.ResolveUnixAddr("unix", path)
				if err != nil {
					return err
				}
				conn, err := net.DialUnix("unix", nil, addr)
				if err != nil {
					return err
				}
				order, orderErr = detectByteOrder(conn)
				conn.Close()
				i3ln.Close() // unblock Accept and return an error
				return orderErr
			})
			eg.Go(func() error {
				for {
					conn, err := i3ln.Accept()
					if err != nil {
						return err
					}
					eg.Go(func() error {
						defer conn.Close()
						for {
							var request [14 + 65536 + 256]byte
							if _, err := io.ReadFull(conn, request[:]); err != nil {
								return err
							}
							if reply := protocol[string(request[:])]; reply != nil {
								if _, err := io.Copy(conn, bytes.NewReader(reply)); err != nil {
									return err
								}
								continue
							}
							// silently drop unexpected messages like i3
						}
					})
				}
			})
			if err := eg.Wait(); err != nil {
				// If order != nil && orderErr == nil, the test succeeded and any
				// returned errors are from teardown.
				if order == nil || orderErr != nil {
					t.Fatal(err)
				}
			}
			if got, want := order, i3order; got != want {
				t.Fatalf("unexpected byte order: got %v, want %v", got, want)
			}
		})
	}
}
