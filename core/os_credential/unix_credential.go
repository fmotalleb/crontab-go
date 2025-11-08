//go:build unix

// Package credential provides functions to switch group and user for command execution.
package credential

import (
	"errors"
	"fmt"
	"os/exec"
	osUser "os/user"
	"strconv"
	"syscall"

	"go.uber.org/zap"
)

func Validate(log *zap.Logger, usr string, grp string) error {
	cu, err := osUser.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user error: %w", err)
	}
	if usr != "" && cu.Uid != "0" {
		return errors.New("cannot switch user of tasks without root privilege, if you need to use user in tasks run crontab-go as user root")
	}
	_, _, err = lookupUIDAndGID(usr, log)
	if err != nil {
		return fmt.Errorf("cannot get uid and gid of user `%s` error: %w", usr, err)
	}
	_, err = lookupGID(grp, log)
	if err != nil {
		return fmt.Errorf("cannot get gid of group `%s` error: %w", grp, err)
	}

	return nil
}

func SetUser(log *zap.Logger, proc *exec.Cmd, usr string, grp string) {
	if usr == "" {
		log.Debug("no username given, running as current user")
		return
	}

	uid, gid, err := lookupUIDAndGID(usr, log)
	if err != nil {
		log.Panic("cannot get uid and gid of user", zap.String("user", usr), zap.Error(err))
	}
	if grp != "" {
		gid, _ = lookupGID(grp, log)
	}

	setUID(proc, uid, gid)
}

func lookupGID(grp string, log *zap.Logger) (gid uint32, err error) {
	if grp == "" {
		return 0, nil
	}
	g, err := osUser.LookupGroup(grp)
	if err != nil {
		log.Panic("cannot find group", zap.String("group", grp), zap.Error(err))
	}
	gidU, err := strconv.ParseUint(g.Gid, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(gidU), nil
}

func lookupUIDAndGID(usr string, log *zap.Logger) (uid uint32, gid uint32, err error) {
	if usr == "" {
		return 0, 0, nil
	}
	u, err := osUser.Lookup(usr)
	if err != nil {
		log.Panic("cannot get uid and gid of user", zap.String("user", usr), zap.Error(err))
	}
	uidU, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	gidU, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return uint32(uidU), uint32(gidU), nil
}

func setUID(
	proc *exec.Cmd,
	uid uint32,
	gid uint32,
) {
	attrib := &syscall.SysProcAttr{}
	proc.SysProcAttr = attrib
	proc.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}
}
