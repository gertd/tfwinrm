package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"log"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	fmt.Println("tfcomm")
	t0 := time.Now()

	state := terraform.InstanceState{
		Ephemeral: terraform.EphemeralState{
			ConnInfo: map[string]string{
				"type":     `winrm`,
				"user":     `Administrator`,
				"password": `verysecure`,
				"host":     `ipaddress`,
				"insecure": `false`,
			},
		},
	}

	comm, err := communicator.New(&state)
	if err != nil {
		log.Fatalln(err)
	}

	var fspew = DebugSpewFunc(spew)

	err = comm.Connect(fspew)
	if err != nil {
		log.Fatalln(err)
	}

	// UploadDir(dst, src)
	// err = comm.UploadDir(".", ".\facter")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// Upload
	f, err := os.Open("./facter/facter.zip")
	if err != nil {
		log.Fatalln(err)
	}

	err = comm.Upload("./facter.zip", f)
	if err != nil {
		log.Fatalln(err)
	}

	if err := execCmd(comm, `powershell.exe -nologo -noprofile -command "& { Add-Type -A 'System.IO.Compression.FileSystem'; [IO.Compression.ZipFile]::ExtractToDirectory('./facter.zip', './facter'); }"`); err != nil {
		log.Fatalln(err)
	}
	if err := execCmd(comm, `del .\facter.zip`); err != nil {
		log.Fatalln(err)
	}
	if err := execCmd(comm, `.\facter\facter.exe`); err != nil {
		log.Fatalln(err)
	}
	if err := execCmd(comm, `rmdir /s /q facter`); err != nil {
		log.Fatalln(err)
	}

	comm.Disconnect()

	t1 := time.Now()
	fmt.Printf("Duration %v\n", t1.Sub(t0))
}

func execCmd(comm communicator.Communicator, cmdStmt string) error {

	// output buffers for SSH
	var outBuf, errBuf bytes.Buffer

	cmd := remote.Cmd{
		Command: cmdStmt,
		Stdin:   os.Stdin,
		Stdout:  &outBuf,
		Stderr:  &errBuf,
	}

	err := comm.Start(&cmd)
	if err != nil {
		return err
	}
	cmd.Wait()

	fmt.Println(outBuf.String())

	return nil
}

func spew(msg string) {
	log.Println(msg)
}

// DebugSpewFunc --
type DebugSpewFunc func(string)

// Output -- spew implementation function
func (f DebugSpewFunc) Output(msg string) {
	f(msg)
}
