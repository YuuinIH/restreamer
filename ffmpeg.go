package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"sync"
)

func NewFFmpeg() (*ffmpeg, error) {
	return &ffmpeg{
		cmd:    &exec.Cmd{},
		stdIn:  nil,
		stdout: nil,
		close:  new(sync.Once),
		debug:  true,
	}, nil
}

type ffmpeg struct {
	cmd    *exec.Cmd
	stdIn  io.WriteCloser
	stdout io.ReadCloser
	close  *sync.Once
	debug  bool
}

func (r *ffmpeg) Stream(sourceurl url.URL, streamurl url.URL) error {
	r.cmd = exec.Command("ffmpeg", "-re", "-i", sourceurl.String(), "-fflags", "nobuffer", "-strict", "experimental", "-avioflags", "direct", "-fflags", "discardcorrupt", "-analyzeduration", "1000000", "-flags", "low_delay",
		"-rtsp_transport", "tcp", "-c", "copy", "-f", "flv", "-g", "5", streamurl.String())
	var err error
	if r.stdIn, err = r.cmd.StdinPipe(); err != nil {
		return err
	}
	if r.stdout, err = r.cmd.StdoutPipe(); err != nil {
		return err
	}
	if r.debug {
		r.cmd.Stderr = os.Stderr
	}
	fmt.Print(streamurl.String())
	err = r.cmd.Start()
	if err != nil {
		r.cmd.Process.Kill()
		return err
	}
	return r.cmd.Wait()
}

func (r *ffmpeg) Stop() error {
	r.close.Do(
		func() {
			if r.cmd.ProcessState == nil {
				r.stdIn.Write([]byte("q"))
			}
		})
	return nil
}
