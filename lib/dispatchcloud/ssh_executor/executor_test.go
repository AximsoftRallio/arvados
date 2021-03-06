// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package ssh_executor

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"git.curoverse.com/arvados.git/lib/dispatchcloud/test"
	"golang.org/x/crypto/ssh"
	check "gopkg.in/check.v1"
)

// Gocheck boilerplate
func Test(t *testing.T) {
	check.TestingT(t)
}

var _ = check.Suite(&ExecutorSuite{})

type testTarget struct {
	test.SSHService
}

func (*testTarget) VerifyHostKey(ssh.PublicKey, *ssh.Client) error {
	return nil
}

type ExecutorSuite struct{}

func (s *ExecutorSuite) TestExecute(c *check.C) {
	command := `foo 'bar' "baz"`
	stdinData := "foobar\nbaz\n"
	_, hostpriv := test.LoadTestKey(c, "../test/sshkey_vm")
	clientpub, clientpriv := test.LoadTestKey(c, "../test/sshkey_dispatch")
	for _, exitcode := range []int{0, 1, 2} {
		srv := &testTarget{
			SSHService: test.SSHService{
				Exec: func(cmd string, stdin io.Reader, stdout, stderr io.Writer) uint32 {
					c.Check(cmd, check.Equals, command)
					var wg sync.WaitGroup
					wg.Add(2)
					go func() {
						io.WriteString(stdout, "stdout\n")
						wg.Done()
					}()
					go func() {
						io.WriteString(stderr, "stderr\n")
						wg.Done()
					}()
					buf, err := ioutil.ReadAll(stdin)
					wg.Wait()
					c.Check(err, check.IsNil)
					if err != nil {
						return 99
					}
					_, err = stdout.Write(buf)
					c.Check(err, check.IsNil)
					return uint32(exitcode)
				},
				HostKey:        hostpriv,
				AuthorizedKeys: []ssh.PublicKey{clientpub},
			},
		}
		err := srv.Start()
		c.Check(err, check.IsNil)
		c.Logf("srv address %q", srv.Address())
		defer srv.Close()

		exr := New(srv)
		exr.SetSigners(clientpriv)

		done := make(chan bool)
		go func() {
			stdout, stderr, err := exr.Execute(command, bytes.NewBufferString(stdinData))
			if exitcode == 0 {
				c.Check(err, check.IsNil)
			} else {
				c.Check(err, check.NotNil)
				err, ok := err.(*ssh.ExitError)
				c.Assert(ok, check.Equals, true)
				c.Check(err.ExitStatus(), check.Equals, exitcode)
			}
			c.Check(stdout, check.DeepEquals, []byte("stdout\n"+stdinData))
			c.Check(stderr, check.DeepEquals, []byte("stderr\n"))
			close(done)
		}()

		timeout := time.NewTimer(time.Second)
		select {
		case <-done:
		case <-timeout.C:
			c.Fatal("timed out")
		}
	}
}
