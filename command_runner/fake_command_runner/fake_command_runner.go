package fake_command_runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"sync"
)

type FakeCommandRunner struct {
	executedCommands     []*exec.Cmd
	startedCommands      []*exec.Cmd
	backgroundedCommands []*exec.Cmd
	waitedCommands       []*exec.Cmd
	killedCommands       []*exec.Cmd
	signalledCommands    map[*exec.Cmd]os.Signal

	commandCallbacks map[*CommandSpec]func(*exec.Cmd) error
	waitingCallbacks map[*CommandSpec]func(*exec.Cmd) error

	process *os.Process

	sync.RWMutex
}

type CommandSpec struct {
	Path  string
	Args  []string
	Env   []string
	Stdin string
	Dir   string
}

func (s CommandSpec) Matches(cmd *exec.Cmd) bool {
	if s.Path != "" && !(s.Path == cmd.Path || s.Path == filepath.Base(cmd.Path)) {
		return false
	}

	if s.Dir != "" && s.Dir != cmd.Dir {
		return false
	}

	if len(s.Args) > 0 && !reflect.DeepEqual(s.Args, cmd.Args[1:]) {
		return false
	}

	if len(s.Env) > 0 && !reflect.DeepEqual(s.Env, cmd.Env) {
		return false
	}

	if s.Stdin != "" {
		if cmd.Stdin == nil {
			return false
		}

		in := make([]byte, len(s.Stdin))
		_, err := cmd.Stdin.Read(in)
		if err != nil {
			return false
		}

		if string(in) != s.Stdin {
			return false
		}
	}

	return true
}

func New() *FakeCommandRunner {
	return &FakeCommandRunner{
		signalledCommands: make(map[*exec.Cmd]os.Signal),

		commandCallbacks: make(map[*CommandSpec]func(*exec.Cmd) error),
		waitingCallbacks: make(map[*CommandSpec]func(*exec.Cmd) error),
	}
}

func (r *FakeCommandRunner) Run(cmd *exec.Cmd) error {
	r.RLock()
	callbacks := r.commandCallbacks
	r.RUnlock()

	r.Lock()
	r.executedCommands = append(r.executedCommands, cmd)
	r.Unlock()

	for spec, callback := range callbacks {
		if spec.Matches(cmd) {
			return callback(cmd)
		}
	}

	r.RLock()
	if r.process != nil {
		cmd.Process = r.process
	}
	r.RUnlock()

	return nil
}

func (r *FakeCommandRunner) RunInjectsProcessToCmd(process *os.Process) {
	r.Lock()
	r.process = process
	r.Unlock()
}

func (r *FakeCommandRunner) Start(cmd *exec.Cmd) error {
	r.RLock()
	callbacks := r.commandCallbacks
	r.RUnlock()

	r.Lock()
	r.startedCommands = append(r.startedCommands, cmd)
	r.Unlock()

	for spec, callback := range callbacks {
		if spec.Matches(cmd) {
			return callback(cmd)
		}
	}

	r.RLock()
	if r.process != nil {
		cmd.Process = r.process
	}
	r.RUnlock()

	return nil
}

func (r *FakeCommandRunner) Background(cmd *exec.Cmd) error {
	r.RLock()
	callbacks := r.commandCallbacks
	r.RUnlock()

	r.Lock()
	r.backgroundedCommands = append(r.backgroundedCommands, cmd)
	r.Unlock()

	for spec, callback := range callbacks {
		if spec.Matches(cmd) {
			return callback(cmd)
		}
	}

	return nil
}

func (r *FakeCommandRunner) Wait(cmd *exec.Cmd) error {
	r.RLock()
	callbacks := r.waitingCallbacks
	r.RUnlock()

	r.Lock()
	r.waitedCommands = append(r.waitedCommands, cmd)
	r.Unlock()

	for spec, callback := range callbacks {
		if spec.Matches(cmd) {
			return callback(cmd)
		}
	}

	return nil
}

func (r *FakeCommandRunner) Kill(cmd *exec.Cmd) error {
	r.Lock()
	defer r.Unlock()

	r.killedCommands = append(r.killedCommands, cmd)

	return nil
}

func (r *FakeCommandRunner) Signal(cmd *exec.Cmd, signal os.Signal) error {
	r.Lock()
	defer r.Unlock()

	r.signalledCommands[cmd] = signal

	return nil
}

func (r *FakeCommandRunner) WhenRunning(spec CommandSpec, callback func(*exec.Cmd) error) {
	r.Lock()
	defer r.Unlock()

	r.commandCallbacks[&spec] = callback
}

func (r *FakeCommandRunner) WhenWaitingFor(spec CommandSpec, callback func(*exec.Cmd) error) {
	r.Lock()
	defer r.Unlock()

	r.waitingCallbacks[&spec] = callback
}

func (r *FakeCommandRunner) ExecutedCommands() []*exec.Cmd {
	r.RLock()
	defer r.RUnlock()

	return r.executedCommands
}

func (r *FakeCommandRunner) StartedCommands() []*exec.Cmd {
	r.RLock()
	defer r.RUnlock()

	return r.startedCommands
}

func (r *FakeCommandRunner) BackgroundedCommands() []*exec.Cmd {
	r.RLock()
	defer r.RUnlock()

	return r.backgroundedCommands
}

func (r *FakeCommandRunner) KilledCommands() []*exec.Cmd {
	r.RLock()
	defer r.RUnlock()

	return r.killedCommands
}

func (r *FakeCommandRunner) WaitedCommands() []*exec.Cmd {
	r.RLock()
	defer r.RUnlock()

	return r.waitedCommands
}

func (r *FakeCommandRunner) SignalledCommands() map[*exec.Cmd]os.Signal {
	r.RLock()
	defer r.RUnlock()

	return r.signalledCommands
}
