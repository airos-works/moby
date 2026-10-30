package daemon

import (
	"bytes"
	"context"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/events"
	libcontainerdtypes "github.com/moby/moby/v2/daemon/internal/libcontainerd/types"
	"github.com/moby/moby/v2/errdefs"
	"github.com/pkg/errors"
)

// ContainerTop lists the processes running inside of the given container.
// Haiku's ps has no Linux options (no -ef, no pid selection, no PID column),
// so it runs with its defaults and the container's teams (from the shim,
// which reports the whole tree) are picked out of its output.
func (daemon *Daemon) ContainerTop(name string, psArgs string) (*container.TopResponse, error) {
	if psArgs != "" && psArgs != "-ef" {
		return nil, errdefs.InvalidParameter(errors.New("ps arguments are not supported on this platform"))
	}

	ctr, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	tsk, err := func() (libcontainerdtypes.Task, error) {
		ctr.Lock()
		defer ctr.Unlock()

		tsk, err := ctr.GetRunningTask()
		if err != nil {
			return nil, err
		}
		if ctr.State.Restarting {
			return nil, errContainerIsRestarting(ctr.ID)
		}
		return tsk, nil
	}()
	if err != nil {
		return nil, err
	}

	infos, err := tsk.Pids(context.Background())
	if err != nil {
		return nil, err
	}
	procs := make(map[int]bool, len(infos))
	for _, p := range infos {
		procs[int(p.Pid)] = true
	}

	if isLinuxVMContainer(ctr) {
		procList, err := linuxVMTop(procs)
		if err != nil {
			return nil, errdefs.System(err)
		}
		daemon.LogContainerEvent(ctr, events.ActionTop)
		return procList, nil
	}

	output, err := exec.Command("ps").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			// first line of stderr shows why ps failed
			line := bytes.SplitN(ee.Stderr, []byte{'\n'}, 2)
			if len(line) > 0 && len(line[0]) > 0 {
				err = errors.New(string(line[0]))
			}
		}
		return nil, errdefs.System(errors.Wrap(err, "ps"))
	}
	procList, err := parseHaikuPSOutput(output, procs)
	if err != nil {
		return nil, err
	}
	daemon.LogContainerEvent(ctr, events.ActionTop)
	return procList, nil
}

// parseHaikuPSOutput reads the output of Haiku's ps, whose columns are
// "Team Id #Threads Gid Uid": the team's command line (which may hold
// spaces) and then four numbers, so each line is read from the right.
func parseHaikuPSOutput(output []byte, procs map[int]bool) (*container.TopResponse, error) {
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "Team") {
		return nil, errors.New("Couldn't find the Team header in ps output")
	}
	procList := &container.TopResponse{
		Titles: []string{"UID", "GID", "PID", "THREADS", "CMD"},
	}
	for _, line := range lines[1:] {
		fields := fieldsASCII(line)
		if len(fields) < 5 {
			continue
		}
		n := len(fields)
		pid, err := strconv.Atoi(fields[n-4])
		if err != nil {
			return nil, errors.Errorf("Unexpected pid '%s': %s", fields[n-4], err)
		}
		if !procs[pid] {
			continue
		}
		procList.Processes = append(procList.Processes, []string{
			fields[n-1], fields[n-2], fields[n-4], fields[n-3],
			strings.Join(fields[:n-4], " "),
		})
	}
	return procList, nil
}

// fieldsASCII is similar to strings.Fields but only allows ASCII whitespaces
func fieldsASCII(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case '\t', '\n', '\f', '\r', ' ':
			return true
		}
		return false
	})
}

// linuxVMSocket is containerd-linuxvm's socket (containerd's pkg/linuxvm).
const linuxVMSocket = "/var/run/containerd-linuxvm.sock"

// linuxVMTop lists a Linux container's processes from the VM's ps: procs
// are its pids in the VM, as the shim reports them.
func linuxVMTop(procs map[int]bool) (*container.TopResponse, error) {
	c, err := net.DialTimeout("unix", linuxVMSocket, 5*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "the Linux VM")
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(30 * time.Second))
	if _, err := io.WriteString(c, "ps\n"); err != nil {
		return nil, err
	}
	output, err := io.ReadAll(c)
	if err != nil {
		return nil, err
	}
	status, output, _ := bytes.Cut(output, []byte{'\n'})
	if string(status) != "ok" {
		return nil, errors.Errorf("the Linux VM: %s", status)
	}
	// PID USER PPID STAT TIME COMMAND
	procList := &container.TopResponse{Titles: []string{"UID", "PID", "PPID", "STAT", "TIME", "CMD"}}
	for _, line := range strings.Split(string(output), "\n")[1:] {
		fields := fieldsASCII(line)
		if len(fields) < 6 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil || !procs[pid] {
			continue
		}
		procList.Processes = append(procList.Processes, []string{
			fields[1], fields[0], fields[2], fields[3], fields[4], strings.Join(fields[5:], " "),
		})
	}
	return procList, nil
}
