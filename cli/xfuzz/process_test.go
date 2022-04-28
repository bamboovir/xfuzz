package xfuzz

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"text/template"
	"time"

	"github.com/google/shlex"
	"github.com/google/uuid"
)

type CMDTemplateArgs struct {
	Root string
	File string
}

func NewCMDTemplateArgs(root string) *CMDTemplateArgs {
	args := &CMDTemplateArgs{}
	args.File = filepath.Join(root, uuid.New().String())
	return args
}

var (
	CommandFlag            = flag.String("cmd", "", "command")
	CommandTemplateFlag    = flag.Bool("cmd-template", false, "{{.File}}")
	CorpusFlag             = flag.String("corpus", "", "corpus dir path")
	TmpFlag                = flag.String("tmp", "", "tmp storage dir path")
	SignalFlag             = flag.String("signal", ".*panic.*", "regex exception predicate")
	DeadlineFlag           = flag.Int("deadline", 1, "maximum amount of time a command could exec (second)")
	ExceptionPredicateDict = []string{
		".*panic.*",
		".*core dump.*",
		".*Segmentation fault.*",
		".*segmentation violation.*",
		".*invalid memory.*",
		".*nil pointer dereference.*",
		".*signal SIGSEGV.*",
		".*fatal.*",
	}
)

func LoadCorpus(path string, f *testing.F) (err error) {
	f.Logf("load corpus from [%s]...", path)
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		f.Add(data)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type CMDParser struct {
	t *template.Template
}

func NewCMDParser(cmdText string) (cmd *CMDParser, err error) {
	cmd = &CMDParser{}
	t, err := template.New("cmd").Parse(cmdText)
	if err != nil {
		return nil, err
	}
	cmd.t = t
	return cmd, nil
}

func (c *CMDParser) ParseCMD(cmdArgs *CMDTemplateArgs) (cmd string, args []string, err error) {
	cmdBuf := bytes.NewBuffer(make([]byte, 0))
	err = c.t.Execute(cmdBuf, cmdArgs)
	if err != nil {
		return "", []string{}, err
	}
	args, err = shlex.Split(cmdBuf.String())
	if err != nil {
		return "", []string{}, err
	}
	if len(args) == 0 {
		return "", []string{}, fmt.Errorf("parse command flag failed with zero args")
	}
	return args[0], args[1:], nil
}

func CheckArgs() (err error) {
	if *CommandFlag == "" {
		flag.PrintDefaults()
		return fmt.Errorf("cmd flag is required")
	}

	if *CommandTemplateFlag {
		if *TmpFlag == "" {
			flag.PrintDefaults()
			return fmt.Errorf("tmp flag is required")
		}
	}
	return nil
}

func FuzzProcess(f *testing.F) {
	err := CheckArgs()
	if err != nil {
		flag.PrintDefaults()
		f.Errorf("check args failed with [%s]\n", err)
		return
	}

	if *CorpusFlag != "" {
		err := LoadCorpus(*CorpusFlag, f)
		if err != nil {
			f.Errorf("load corpus from %s failed with [%s]\n", *CorpusFlag, err)
			return
		}
	} else {
		f.Logf("no corpus was declared\n")
	}

	c, err := NewCMDParser(*CommandFlag)
	if err != nil {
		f.Errorf("parse command flag failed with [%s]\n", err)
		return
	}

	if *CommandTemplateFlag {
		err = os.MkdirAll(*TmpFlag, os.ModePerm)
		if err != nil {
			f.Errorf("create tmp dir failed with [%s]\n", err)
			return
		}
	}

	ExceptionPredicateDict = append(ExceptionPredicateDict, *SignalFlag)

	f.Fuzz(func(t *testing.T, data []byte) {
		targs := NewCMDTemplateArgs(*TmpFlag)
		if *CommandTemplateFlag {
			err = os.WriteFile(targs.File, data, 0666)
			if err != nil {
				t.Errorf("write payload file failed with [%s]\n", err)
				return
			}
		}
		name, args, err := c.ParseCMD(targs)
		if err != nil {
			t.Errorf("parse command failed with [%s]\n", err)
			return
		}
		cmd := exec.Command(name, args...)
		cmd.Stdin = bytes.NewReader(data)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Errorf("cmd [%s] failed with [%s]\n", cmd.String(), err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			t.Errorf("cmd [%s] failed with [%s]\n", cmd.String(), err)
			return
		}
		deadline := time.Second * time.Duration(*DeadlineFlag)
		err = cmd.Start()
		t.Logf("cmd [%s] start", cmd.String())
		if err != nil {
			t.Errorf("cmd [%s] failed with [%s]\n", cmd.String(), err)
			return
		}

		outScan := bufio.NewScanner(stdout)
		errScan := bufio.NewScanner(stderr)
		outChan := make(chan error)
		errChan := make(chan error)

		go func() {
			for outScan.Scan() {
				line := outScan.Text()
				for _, pattern := range ExceptionPredicateDict {
					matchOut, err := regexp.MatchString(pattern, line)
					if err != nil {
						outChan <- fmt.Errorf("stdout: [%s] with [%s]", line, err)
						return
					}
					if matchOut {
						outChan <- fmt.Errorf("stdout: [%s]", line)
						return
					}
				}
			}
			outChan <- nil
		}()

		go func() {
			for errScan.Scan() {
				line := errScan.Text()
				for _, pattern := range ExceptionPredicateDict {
					matchErr, err := regexp.MatchString(pattern, line)
					if err != nil {
						errChan <- fmt.Errorf("stderr: [%s] with [%s]", line, err)
						return
					}
					if matchErr {
						errChan <- fmt.Errorf("stderr: [%s]", line)
						return
					}
				}
			}
			errChan <- nil
		}()

		done := make(chan struct{})
		go func() {
			cmd.Wait()
			done <- struct{}{}
			t.Logf("process terminated")
		}()

		terminate := func() {
			cmd.Process.Signal(syscall.SIGINT)
			select {
			case <-time.After(time.Second):
				t.Logf("process not terminated after SIGINT, try SIGKILL")
				cmd.Process.Signal(syscall.SIGTERM)
				select {
				case <-time.After(time.Second):
					t.Logf("process not terminated after SIGTERM, try SIGKILL")
					cmd.Process.Kill()
					select {
					case <-time.After(time.Second):
						t.Logf("process not terminated after SIGKILL, try kill manually")
					case <-done:
						t.Logf("process terminated after SIGKILL")
					}
				case <-done:
				}
			case <-done:
			}
		}

		select {
		case <-time.After(deadline):
			terminate()
		case err := <-outChan:
			if err != nil {
				terminate()
				t.Errorf("stdout: %s", err)
			}
		case err := <-errChan:
			if err != nil {
				terminate()
				t.Errorf("stderr: %s", err)
			}
		case <-done:
		}
		return
	})
}
